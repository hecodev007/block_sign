package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"sync"
	"time"
	"xorm.io/builder"
)

var feeFloat decimal.Decimal
var toCollect decimal.Decimal

type CollectZvcJob struct {
	coinName   string
	cfg        conf.Collect2
	decimalBit int
}

func NewCollectZvcJob(cfg conf.Collect2) cron.Job {
	if cfg.MinAmount < 0.001 {
		panic("zvc最小归集值过小")
	}
	feeFloat = decimal.NewFromFloat(0.01) //预扣手续费
	toCollect, _ = decimal.NewFromString(FlagCollect)

	return CollectZvcJob{
		coinName:   "zvc",
		cfg:        cfg,
		decimalBit: 9,
	}
}

func (c CollectZvcJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	start := time.Now()

	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
	if len(c.cfg.Mchs) != 0 {
		mchs, err = entity.FcMch{}.Find(builder.In("platform", c.cfg.Mchs).And(builder.Eq{"status": 2}))
	} else {
		mchs, err = entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
			Where(builder.Eq{
				"status":    0,
				"coin_name": c.coinName,
			})).And(builder.Eq{"status": 2}))
	}
	if err != nil {
		log.Errorf("find platforms err %v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, tmp := range mchs {
		go func(mch *entity.FcMch) {
			wg.Add(1)
			defer wg.Done()

			if err := c.collect(mch.Id, mch.Platform); err != nil {
				log.Errorf(" %s ## collect err: %v", mch.Platform, err)
			}
		}(tmp)
	}
	wg.Wait()
}

//依赖数据库的方式自动归集
func (c *CollectZvcJob) collect(mchId int, mchName string) error {
	var (
		toAmountOfMin decimal.Decimal
	)
	//默认附加手续费
	toAmountOfMin = decimal.NewFromFloat(c.cfg.MinAmount).Add(feeFloat)
	if toAmountOfMin.LessThan(decimal.Zero) {
		return fmt.Errorf("设置归集的金额过小,[%s]", toAmountOfMin.String())
	}

	//获取归集的目标冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   c.coinName,
	})
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't hava cold address", mchName)
	}

	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >=  ? and forzen_amount = 0", toAmountOfMin.String())).
		And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return fmt.Errorf("%s don't hava need collected address", mchName)
	}

	//生成归集订单
	cltApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   c.coinName,
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mchName,
		Operator:   "Robot",
		AppId:      mchId,
		Type:       "gj",
		Purpose:    "自动归集",
		Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}

	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	for _, to := range toAddrs {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     to,
			AddressFlag: "to",
			Status:      0,
		})
	}
	for _, from := range fromAddrs {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from.Address,
			AddressFlag: "from",
			Status:      0,
			ToAmount:    from.Amount,
		})
	}

	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		return err
	}
	//开始请求钱包服务归集
	for i, from := range fromAddrs {
		//随机获取冷地址
		to := toAddrs[i%len(toAddrs)]
		orderReq := &transfer.ZvcOrderRequest{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = cltApply.OutOrderid
		orderReq.OrderNo = fmt.Sprintf("%s_%d", cltApply.OrderId, i)
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = c.coinName
		orderReq.FromAddress = from.Address
		orderReq.ToAddress = to
		orderReq.ToAmount = toCollect
		//直接发起交易
		_, err := c.walletServerCreate(orderReq)
		if err != nil {
			log.Errorf("%s 归集交易失败，%s", c.coinName, err.Error())
			//更新减少冻结金额
			continue
		}
		log.Infof("address：%s,归集交易成功", from)
	}
	return nil
}

func (c *CollectZvcJob) transfer(fc *entity.FcCollect) (txid string, err error) {
	orderReq := new(transfer.ZvcOrderRequest)
	err = json.Unmarshal([]byte(fc.SendData), orderReq)
	if err != nil {
		return "", err
	}
	txid, err = c.walletServerCreate(orderReq)
	if err != nil {
		//修改为失败
		return "", err
	}
	if txid == "" {
		return "", errors.New("empty txid")
	}
	return txid, nil
}

//=======================私有方法==========================
func (c *CollectZvcJob) buildOrder(toAmount, fee decimal.Decimal, fa *entity.FcAddressAmount) (*transfer.ZvcOrderRequest, error) {
	//账户模型没有找零
	//私有方法 构建cocos订单
	var (
		fromAddr string
		toAddr   string
		//changeAddr string
	)
	// 查找from地址和金额
	fromAddr = fa.Address

	////查询这个币种的找零地址
	//changeAddrs, err := dao.FcGenerateAddressListFindChangeAddr(int(fa.AppId), fa.CoinType)
	//if err != nil {
	//	return nil, fmt.Errorf("商户归集异常,无法查询找零地址，商户：%d,coinName:%s,err:%s", fa.AppId, fa.CoinType, err.Error())
	//}
	//随机选一个
	//index := util.RandInt64(0, int64(len(changeAddrs)))
	//changeAddr = changeAddrs[index]

	//查询这个币种的归集地址
	toAddrs, err := dao.FcGenerateAddressListFindAddresses(1, 2, int(fa.AppId), fa.CoinType)
	if err != nil {
		return nil, fmt.Errorf("商户归集异常,无法查询归集地址，商户：%d,coinName:%s,err:%s", fa.AppId, fa.CoinType, err.Error())

	}
	//随机选一个
	index := util.RandInt64(0, int64(len(toAddrs)))
	toAddr = toAddrs[index]

	outerOrderNo := util.GetUUID()
	//填充参数
	orderReq := &transfer.ZvcOrderRequest{}
	orderReq.ApplyId = -1
	orderReq.OuterOrderNo = outerOrderNo
	orderReq.OrderNo = outerOrderNo
	orderReq.MchName = "robot"
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = fa.CoinType
	orderReq.ToAddress = toAddr
	orderReq.Memo = outerOrderNo
	orderReq.ToAmount = toAmount
	return orderReq, nil
}

//创建交易接口参数
func (c *CollectZvcJob) walletServerCreate(orderReq *transfer.ZvcOrderRequest) (txid string, err error) {
	dd, _ := json.Marshal(orderReq)
	log.Infof("zvc 交易发送内容 :%s", string(dd))
	data, err := util.PostJsonByAuth(c.cfg.Url+"/v1/zvc/Transfer", c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("zvc 交易返回内容 :%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败:%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常,%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}
