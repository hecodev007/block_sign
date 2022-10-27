package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type FeeBNBToColdJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewFeeBNBToColdJob(cfg conf.Collect2) cron.Job {
	return FeeBNBToColdJob{
		coinName: "bnb",
		cfg:      cfg,
	}
}

func (c FeeBNBToColdJob) Run() {
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
			if err := c.collect(mch.Id, mch.Platform, "bnb1u8cg55dw6ls8z5ht3sezpu4jm09lu6t9dm34qy"); err != nil {
				log.Errorf(" %s ## collect err: %v", mch.Platform, err)
			}
		}(tmp)
	}
	wg.Wait()
}

func (c *FeeBNBToColdJob) collect(mchId int, mchName, toAddr string) error {

	//获取手续费地址
	changeAddr, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 3, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >=  ? and forzen_amount = 0", c.cfg.NeedFee)), 1)
	if err != nil {
		log.Errorf("查询手续地址异常:%s", err.Error())
		return err
	}
	if len(changeAddr) == 0 {
		log.Errorf("%s don't hava need collected address", mchName)
		return fmt.Errorf("%s don't hava need collected address", mchName)
	}

	//生成归集订单
	cltApply := &entity.FcTransfersApply{
		Username:   "Robot",
		Department: "blockchains-go",
		Applicant:  mchName,
		OutOrderid: fmt.Sprintf("FEE_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Operator:   "Robot",
		CoinName:   c.coinName,
		Type:       "fee",
		Purpose:    fmt.Sprintf("%s手续费", c.coinName),
		Lastmodify: util.GetChinaTimeNow(),
		AppId:      mchId,
		Source:     1,
		Status:     int(entity.ApplyStatus_Fee), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
	}
	amount := decimal.NewFromFloat(c.cfg.NeedFee)
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     toAddr,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     changeAddr[0].Address,
		AddressFlag: "from",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err == nil {
		//开始请求钱包服务归集
		orderReq := &transfer.BNBOrderRequest{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = cltApply.OutOrderid
		orderReq.OrderNo = cltApply.OrderId
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = c.coinName
		orderReq.FromAddress = changeAddr[0].Address
		orderReq.ToAddress = toAddr
		orderReq.Token = strings.ToUpper(c.coinName)
		orderReq.Quantity = amount.Shift(8).String()
		if err := c.walletServerCreateCold(orderReq); err != nil {
			log.Errorf("err : %v", err)
		} else {
			log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s]",
				strings.ToUpper(c.coinName), mchId, appId, changeAddr[0].Address, toAddr, amount.String())
		}
	}
	return nil
}

func (c *FeeBNBToColdJob) walletServerCreateCold(orderReq *transfer.BNBOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Quantity, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("%s walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.CoinName, orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("%s walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.CoinName, orderReq.OuterOrderNo, string(data))
	}
	return nil
}
