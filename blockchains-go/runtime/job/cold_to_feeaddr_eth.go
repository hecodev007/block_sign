package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"sync"
	"time"
	"xorm.io/builder"
)

//冷地址划转给手续费地址
type ColdToFeeEthJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewColdToFeeEthJob(cfg conf.Collect2) cron.Job {
	return ColdToFeeEthJob{
		coinName: "eth",
		cfg:      cfg,
	}
}

func (c ColdToFeeEthJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	start := time.Now()

	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
	if global.G_TASK_MARK {
		log.Infof("*** %s 任务正在执行，等待下一轮", c.coinName)
	}

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
	global.G_TASK_MARK = true
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
	global.G_TASK_MARK = false
}

func (c *ColdToFeeEthJob) collect(mchId int, mchName string) error {

	coldAddr := "0x0055e75217ca5cb5aa8290cd966f9d85751a7993"
	feeAddr := "0x0000bf7f4e4b7fb2315fc6d5d0f8854c91dff1d8"

	//构建交易
	//生成手续费订单
	feeApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   "eth",
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
		orderReq := &transfer.EthTransferFeeReq{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = feeApply.OutOrderid
		orderReq.OrderNo = feeApply.OrderId
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = "eth"
		orderReq.FromAddr = coldAddr
		orderReq.ToAddrs = []string{feeAddr}
		orderReq.NeedFee = decimal.NewFromFloat(5).Shift(18).String() //eth -> wei
		if err = c.walletServerFee(orderReq); err != nil {
			log.Errorf("[%s] 地址大手续费错误，Err=[%v]", coldAddr, err)
			return err
		}
	}

	return nil
}

//创建交易接口参数
func (c *ColdToFeeEthJob) walletServerFee(orderReq *transfer.EthTransferFeeReq) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/fee", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s fee send :%s", c.coinName, string(dd))
	log.Infof("%s fee resp :%s", c.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerFee 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerFee 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}
