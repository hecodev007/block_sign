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
	"sync"
	"time"
	"xorm.io/builder"
)

//冷地址划转给手续费地址
type ColdToFeeOngJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewColdToFeeOngJob(cfg conf.Collect2) cron.Job {
	return ColdToFeeOngJob{
		coinName: "ong",
		cfg:      cfg,
	}
}

func (c ColdToFeeOngJob) Run() {
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

func (c *ColdToFeeOngJob) collect(mchId int, mchName string) error {

	var txid string

	coldAddr := "AHHNKyeYyVrP15j1Xwd2XRmadR6AhSsmyt"
	feeAddr := "AP6YtupEhzFETidA3X3kwdWHcRMUBq9yPk"

	//构建交易
	//生成手续费订单
	feeApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   "ont",
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("COlD_TO_FEE_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mchName,
		Operator:   "Robot",
		AppId:      mchId,
		Type:       "cold2fee",
		Purpose:    "自动归集",
		Status:     int(entity.ApplyStatus_Fee), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}

	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     coldAddr,
		AddressFlag: "from",
		Status:      0,
	})
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     feeAddr,
		AddressFlag: "to",
		Status:      0,
	})
	appId, err := feeApply.TransactionAdd(applyAddresses)
	if err == nil {
		//开始请求钱包服务归集
		orderReq := &transfer.OntOrderRequest{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = feeApply.OutOrderid
		orderReq.OrderNo = feeApply.OrderId
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = "ong"
		orderReq.FromAddress = coldAddr
		orderReq.ToAddress = feeAddr
		orderReq.Amount = decimal.NewFromFloat(100.0).Shift(9).IntPart() //
		if txid, err = c.walletServerFee(orderReq); err != nil {
			log.Errorf("[%s] 地址大手续费错误，Err=[%v]", coldAddr, err)
			return err
		} else {
			log.Infof("success,txid=[%s]", txid)
		}
	}

	return nil
}

//创建交易接口参数
func (c *ColdToFeeOngJob) walletServerFee(orderReq *transfer.OntOrderRequest) (string, error) {
	log.Infof("path:%s", c.cfg.Url)
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/Transfer", c.cfg.Url, "ont"), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", c.coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", c.coinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.(string), nil
}
