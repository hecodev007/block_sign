package job

import (
	"encoding/json"
	"errors"
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
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectHxJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectHxJob(cfg conf.Collect2) cron.Job {
	//initDingErrBot()
	return CollectHxJob{
		coinName: "hx",
		cfg:      cfg,
	}
}

func (c CollectHxJob) Run() {
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
func (c *CollectHxJob) collect(mchId int, mchName string) error {
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
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >= ? and forzen_amount = 0", c.cfg.MinAmount)).
		And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
	if err != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return fmt.Errorf("%s don't have need collected from address", mchName)
	}
	//获取币种的配置
	hxCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(hxCoins) == 0 {
		return errors.New("do not find hx coin")
	}

	//只有一个coin
	coin := hxCoins[0]
	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	for _, from := range fromAddrs {
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
			Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
			Lastmodify: util.GetChinaTimeNow(),
			AppId:      mchId,
			Source:     1,
			Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
			Createtime: time.Now().Unix(),
		}
		if coin.Name != c.coinName {
			cltApply.Eostoken = coin.Token
			cltApply.Eoskey = coin.Name
		}
		amount, _ := decimal.NewFromString(from.Amount)
		//减去手续费
		amount = amount.Sub(fee)
		applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
		to := toAddrs[0] //随机取个地址
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
			//填充参数
			orderReq := &transfer.HxOrderRequest{}
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(mchId)
			orderReq.MchName = mchName
			orderReq.CoinName = c.coinName
			orderReq.Worker = service.GetWorker(c.coinName)
			orderReq.FromAddress = from.Address
			orderReq.ToAddress = to
			orderReq.Amount = amount.Shift(int32(coin.Decimal)).IntPart()

			//构建订单数组
			var orderAddress []transfer.HxOrderAddress
			orderFrom := transfer.HxOrderAddress{
				Dir:     0,
				Address: from.Address,
				Amount:  0,
			}
			orderAddress = append(orderAddress, orderFrom)
			orderTo := transfer.HxOrderAddress{
				Dir:     1,
				Address: from.Address,
				Amount:  amount.Shift(int32(coin.Decimal)).IntPart(),
			}
			orderAddress = append(orderAddress, orderTo)
			orderReq.OrderAddress = orderAddress

			//发送交易
			if err := c.walletServerCreateCold(orderReq); err != nil {
				log.Errorf("发送交易失败： err : %v", err)
			} else {
				log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s]",
					strings.ToUpper(coin.Name), mchId, appId, from.Address, to, amount.String())
			}
		}
	}
	return nil
}

func (c *CollectHxJob) walletServerCreateCold(orderReq *transfer.HxOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%d],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("%s walletServerCollect 请求下单接口失败，outOrderId：%s：,data=: %s", orderReq.CoinName, orderReq.OuterOrderNo, string(data))
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("%s walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s, data: %s", orderReq.CoinName, orderReq.OuterOrderNo, string(data))
	}
	return nil
}
