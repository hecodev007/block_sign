package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"sync"
	"time"
	"xorm.io/builder"
)

type FeeEthToColdJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewFeeEthToColdJob(cfg conf.Collect2) cron.Job {
	return FeeEthToColdJob{
		coinName: "eth",
		cfg:      cfg,
	}
}

func (c FeeEthToColdJob) Run() {
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

func (c *FeeEthToColdJob) collect(mchId int, mchName string) error {

	coldAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 1, "coin_type": "eth", "app_id": mchId}.
		And(builder.Expr("amount <= 1.1 ")), 10)
	if err != nil {
		return err
	}
	if len(coldAddrs) == 0 {
		fmt.Println("没有查找小于条件金额的冷地址！！！")
		return nil
	}

	//查找手续费地址
	feeAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 3, "coin_type": "eth", "app_id": mchId}.
		And(builder.Expr("amount >= 0.5 and forzen_amount = 0")), 10)
	if err != nil {
		return err
	}
	if len(feeAddrs) == 0 {
		return errors.New("没有查找到手续费地址大于0.5eth的地址！！！")
	}
	feeAddr := feeAddrs[rand.Intn(len(feeAddrs))]
	for _, cold := range coldAddrs {
		//构建交易
		//频率过快，底层出错
		time.Sleep(15 * time.Second)

		//生成手续费订单
		feeApply := &entity.FcTransfersApply{
			Username:   "Robot",
			CoinName:   "eth",
			Department: "blockchains-go",
			OutOrderid: fmt.Sprintf("FEE_%d", time.Now().Nanosecond()),
			OrderId:    util.GetUUID(),
			Applicant:  mchName,
			Operator:   "Robot",
			AppId:      mchId,
			Type:       "fee",
			Purpose:    "自动归集",
			Status:     int(entity.ApplyStatus_Fee), //因为是即时归集，所以直接把状态置为构建成功
			Createtime: time.Now().Unix(),
			Lastmodify: util.GetChinaTimeNow(),
			Source:     1,
		}

		applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     feeAddr.Address,
			AddressFlag: "from",
			Status:      0,
		})
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     cold.Address,
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
			orderReq.FromAddr = feeAddr.Address
			orderReq.ToAddrs = []string{cold.Address}
			//orderReq.NeedFee = decimal.NewFromFloat(0.2).Shift(18).String() //eth -> wei
			orderReq.NeedFee = decimal.NewFromFloat(0.2).Shift(18).String() //eth -> wei
			if err = c.walletServerFee(orderReq); err != nil {
				return err
			}
		}
	}

	return nil
}

//创建交易接口参数
func (c *FeeEthToColdJob) walletServerFee(orderReq *transfer.EthTransferFeeReq) error {
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
