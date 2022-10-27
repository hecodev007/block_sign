package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectCkbJob struct {
	coinName string
	cfg      conf.Collect2
	limitMap sync.Map
}

func NewCollectCkbJob(cfg conf.Collect2) cron.Job {
	return CollectCkbJob{
		coinName: "ckb",
		cfg:      cfg,
		limitMap: sync.Map{}, //初始化限制表
	}
}

func (c CollectCkbJob) Run() {
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

func (c *CollectCkbJob) collect(mchId int, mchName string) error {
	////获取币种的配置
	//ckbCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	//if err != nil {
	//	return err
	//}
	//if len(ckbCoins) == 0 {
	//	return errors.New("do not find dot coin")
	//}
	////只有一个coin
	//coin := ckbCoins[0]
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount > ? and forzen_amount = 0", c.cfg.MinAmount)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil
		//return fmt.Errorf("%s don't have need collected from address", mchName)
	}
	//获取归集的目标冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   c.coinName,
	})
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't have cold address", mchName)
	}
	if mchId == 1 && toAddrs[0] != "ckb1qyq9dvdrhmtxep45q9gl9wsmlc5kqvkaxpasep88kh" {
		return fmt.Errorf("%s err address", mchName)
	}

	to := toAddrs[rand.Intn(len(toAddrs))] //随机取个地址
	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	for _, from := range fromAddrs {
		amount, _ := decimal.NewFromString(from.Amount)

		//减去手续费
		amount = amount.Sub(fee)
		if amount.LessThanOrEqual(decimal.NewFromInt(0)) {
			continue
		}
		//if amount.Cmp(decimal.NewFromInt(50000)) == 1 {
		//	log.Infof("ckb1qyqwg5tlu45dejg0u5u8wy8wd5p879qx3wds9p6laa 调整金额")
		//	amount = decimal.NewFromInt(50000)
		//}
		//生产归集订单
		cltApply := &entity.FcTransfersApply{
			Username:   "Robot",
			Department: "blockchains-go",
			Applicant:  mchName,
			OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
			OrderId:    util.GetUUID(),
			Operator:   "Robot",
			CoinName:   c.coinName,
			Type:       "gj",
			Purpose:    fmt.Sprintf("%s自动归集", strings.ToUpper(c.coinName)),
			Lastmodify: util.GetChinaTimeNow(),
			AppId:      mchId,
			Source:     1,
			Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
			Createtime: time.Now().Unix(),
		}

		applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     to,
			AddressFlag: "to",
			Status:      0,
			Lastmodify:  cltApply.Lastmodify,
		})
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from.Address,
			AddressFlag: "from",
			Status:      0,
			Lastmodify:  cltApply.Lastmodify,
		})
		appId, err := cltApply.TransactionAdd(applyAddresses)
		if err == nil {
			orderReq := new(transfer.CkbOrderRequest)
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(mchId)
			orderReq.MchName = mchName
			orderReq.CoinName = c.coinName
			orderReq.Worker = service.GetWorker(c.coinName)

			var orderAddreses []map[string]interface{}
			fromOrder := map[string]interface{}{
				"dir":     0,
				"address": from.Address,
			}
			toOrder := map[string]interface{}{
				"dir":      1,
				"address":  to,
				"quantity": -8,
			}
			changeOrder := map[string]interface{}{
				"dir":     2,
				"address": to,
			}
			orderAddreses = append(orderAddreses, fromOrder)
			orderAddreses = append(orderAddreses, toOrder)
			orderAddreses = append(orderAddreses, changeOrder)
			orderReq.OrderAddress = orderAddreses
			orderReq.FeeString = "0"
			orderReq.IsForce = true
			err = c.walletServerCreate(orderReq)
			if err != nil {
				log.Errorf("%s归集错误,Address = [%s],Err=%v", mchName, from.Address, err)
			}
		} else {
			log.Errorf("%s 构建订单错误： %v", from.Address, err)
		}
	}
	return nil
}

//创建交易接口参数
func (c *CollectCkbJob) walletServerCreate(orderReq *transfer.CkbOrderRequest) error {
	reqData, _ := json.Marshal(orderReq)
	log.Infof("Collect req data: %s", string(reqData))
	data, err := util.PostJsonByAuth(c.cfg.Url+"/ckb/create", c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {

		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil
}
