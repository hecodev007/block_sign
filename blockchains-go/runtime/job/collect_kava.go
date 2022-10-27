package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectKavaJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectKavaJob(cfg conf.Collect2) cron.Job {
	return CollectKavaJob{
		coinName: "kava",
		cfg:      cfg,
	}
}

func (c CollectKavaJob) Run() {
	start := time.Now()
	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
	mchs, err := entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
		Where(builder.Eq{
			"status":    0,
			"coin_name": c.coinName,
		})).And(builder.Eq{"status": 2}))
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
func (c *CollectKavaJob) collect(mchId int, mchName string) error {
	time.Sleep(1 * time.Second)

	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	//feeInt64 := fee.Shift(6)

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

	coinInfos, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return fmt.Errorf("缺少币种设置 err:%s", c.coinName)
	}
	if len(coinInfos) == 0 {
		return fmt.Errorf("缺少币种设置:%s", c.coinName)
	}
	coinInfo := coinInfos[0]

	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >= ? and forzen_amount = 0", c.cfg.MinAmount)).
		And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return fmt.Errorf("%s don't hava need collected address", mchName)
	}
	//生成归集订单

	//开始请求钱包服务归集
	for i, from := range fromAddrs {
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
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     toAddrs[0],
			AddressFlag: "to",
			Status:      0,
		})
		//for _, to := range toAddrs {
		//	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		//		Address:     to,
		//		AddressFlag: "to",
		//		Status:      0,
		//	})
		//}
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from.Address,
			AddressFlag: "from",
			Status:      0,
		})
		appId, err := cltApply.TransactionAdd(applyAddresses)
		if err != nil {
			//log.Errorf("apply create err : %v", err)
			return err
		}

		fromAm, err := decimal.NewFromString(from.Amount)
		if err != nil {
			log.Errorf("地址：%s", from.Address)
			continue
		}
		am := fromAm.Sub(fee)
		if am.LessThanOrEqual(decimal.Zero) {
			log.Errorf("地址：%s db金额：%s,发送金额", from.Address, fromAm.String(), am.String())
			continue
		}
		//随机获取冷地址
		to := toAddrs[0]
		orderReq := &transfer.KavaOrderRequest{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = cltApply.OutOrderid
		orderReq.OrderNo = fmt.Sprintf("%s_%d", cltApply.OrderId, i)
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = c.coinName
		orderReq.Data = transfer.KavaPaymentRequest{
			FromAddress: from.Address,
			ToAddress:   to,
			//先暂时写死
			Amount: am.Shift(6).IntPart(),
			Memo:   "collect",
			//FeePayer:    to, //归集手续费统一有归集地址出
		}

		//发送交易
		createData, _ := json.Marshal(orderReq)
		orderHot := &entity.FcOrderHot{
			ApplyId:      int(appId),
			ApplyCoinId:  coinInfo.Id,
			OuterOrderNo: cltApply.OutOrderid,
			OrderNo:      cltApply.OrderId,
			MchName:      mchName,
			CoinName:     c.coinName,
			FromAddress:  from.Address,
			ToAddress:    to,
			Amount:       am.Shift(int32(coinInfo.Decimal)).IntPart(), //转换整型
			Quantity:     am.Shift(int32(coinInfo.Decimal)).String(),
			Decimal:      int64(coinInfo.Decimal),
			CreateData:   string(createData),
			Status:       int(status.UnknowErrorStatus),
			CreateAt:     time.Now().Unix(),
			UpdateAt:     time.Now().Unix(),
		}
		//直接发起交易
		txid, err := c.walletServerCreate(orderReq)
		if err != nil {
			log.Errorf("kava 归集交易失败，%s", err.Error())
			//更新减少冻结金额
			continue
		}
		orderHot.Status = int(status.BroadcastStatus)
		// 保存热表
		err = dao.FcOrderHotInsert(orderHot)
		if err != nil {
			err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
			// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
			log.Error(err.Error())
			// 发送给钉钉
			dingding.ErrTransferDingBot.NotifyStr(err.Error())
		}
		log.Infof("address：%s,归集交易成功，txid:%s", from.Address, txid)
	}
	return nil
}

//创建交易接口参数
func (c *CollectKavaJob) walletServerCreate(orderReq *transfer.KavaOrderRequest) (txid string, err error) {
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s walletServerCreate 发送内容 :%s", c.coinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("%s walletServerCreate 返回内容 :%s", c.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("walletServerCreate 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("walletServerCreate 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}
