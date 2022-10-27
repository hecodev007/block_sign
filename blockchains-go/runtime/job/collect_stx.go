package job

import (
	"encoding/json"
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
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectStxJob struct {
	coinName string
	feeName  string
	cfg      conf.Collect2
}

func NewCollectStxJob(cfg conf.Collect2) cron.Job {
	return CollectStxJob{
		coinName: "stx",
		feeName:  "btc-stx",
		cfg:      cfg,
	}
}

func (c CollectStxJob) Run() {
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

func (c *CollectStxJob) collect(mchId int, mchName string) error {
	toAddrs := make([]string, 0)
	var err error
	if mchId == 1 {
		toAddrs = []string{"SP2J02RVJPTKVN81YF9TFPGWEJPHMA79VS5TWGGVK"}
	} else {
		//获取归集的目标冷地址
		toAddrs, err = entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
			"type":        address.AddressTypeCold,
			"status":      address.AddressStatusAlloc,
			"platform_id": mchId,
			"coin_name":   c.coinName,
		})
	}

	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't have cold address", mchName)
	}
	//获取有余额的地址

	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount >= ? ", c.cfg.MinAmount)), c.cfg.MaxCount)
	if err != nil {
		log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err.Error())
	}

	if len(fromAddrs) == 0 {
		log.Infof("%s don't have need collected from address", mchName)
		return nil
	}
	//判断归集地址是否需要打手续费
	var (
		collectAddress []*entity.FcAddressAmount
		needFeeAddress []string
	)
	for _, from := range fromAddrs {
		amount, err := decimal.NewFromString(from.Amount)
		if err != nil {
			continue
		}
		pendingAmount, err := decimal.NewFromString(from.PendingAmount)
		if err != nil {
			continue
		}
		tmpAmount := amount.Sub(pendingAmount)
		if tmpAmount.LessThan(decimal.NewFromFloat(c.cfg.MinAmount)) {
			log.Infof("出账金额不足，Amount=[%s],PendingAmount=[%s],最小出账金额为：[%f]", from.Amount, from.PendingAmount, c.cfg.MinAmount)
			continue
		}
		//转换为 btc 的地址去查询
		addresses, err := dao.FcGenerateAddressFindIn([]string{from.Address})
		if err != nil {
			return err
		}
		if len(addresses) != 1 {
			return fmt.Errorf("查询%sfrom地址失败", c.coinName)
		}
		//randIndex := util.RandInt64(0, int64(len(changes)))
		addr := addresses[0].CompatibleAddress
		isNeed, _ := c.needTransferFee(addr, c.cfg.NeedFee)
		if isNeed {
			//log.Infof("need fee: [%s]",addr)
			needFeeAddress = append(needFeeAddress, addr)
		} else {
			//log.Infof("do not need fee: [%s]",from.Address)
			collectAddress = append(collectAddress, from)
		}
	}
	if len(collectAddress) > 0 {
		//获取币种的配置
		Coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
		if err != nil {
			return err
		}
		if len(Coins) == 0 {
			return fmt.Errorf("do not find %s coin", c.coinName)
		}
		to := toAddrs[0] //限制地址
		changes, err := dao.FcGenerateAddressFindIn([]string{to})
		if err != nil {
			return err
		}
		if len(changes) != 1 {
			return fmt.Errorf("查询%s找零地址失败", c.coinName)
		}
		//随机选择
		//randIndex := util.RandInt64(0, int64(len(changes)))
		//changeAddr := changes[0].CompatibleAddress
		changeAddr := changes[0].Address

		for _, coin := range Coins {
			for _, from := range collectAddress {
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
					Status:     int(entity.ApplyStatus_Merge), // 因为是即时归集，所以直接把状态置为构建成功
					Createtime: time.Now().Unix(),
				}
				if coin.Name != c.coinName {
					cltApply.Eostoken = coin.Token
					cltApply.Eoskey = coin.Name
				}
				amount, _ := decimal.NewFromString(from.Amount)
				forzenAmount, _ := decimal.NewFromString(from.ForzenAmount)
				pendindAmount, _ := decimal.NewFromString(from.PendingAmount)
				//扣除冻结金额
				amount = amount.Sub(forzenAmount)
				amount = amount.Sub(pendindAmount)
				if amount.LessThanOrEqual(decimal.Zero) {
					continue
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
					//填充参数
					orderReq := &transfer.StxOrderRequest{}
					orderReq.ApplyId = appId
					orderReq.OuterOrderNo = cltApply.OutOrderid
					orderReq.OrderNo = cltApply.OrderId
					//orderReq.MchId = int64(mchId)
					orderReq.MchName = mchName
					orderReq.CoinName = strings.ToUpper(c.coinName)
					//orderReq.Worker = service.GetWorker(c.coinName)

					orderReq.FromAddress = from.Address
					orderReq.ChanegeAddress = changeAddr
					orderReq.TransferFee = 0

					orderAddress := transfer.StxToAddrAmount{
						ToAddr:   to,
						ToAmount: amount.Shift(int32(coin.Decimal)).IntPart(),
					}
					orderReq.ToAddrs = append(orderReq.ToAddrs, orderAddress)

					txid, err := c.walletServerCreateHot(orderReq)

					if err != nil {
						log.Errorf(err.Error())
						continue
					}
					log.Infof("%s归集成功，txid=【%s】", c.coinName, txid)
				} else {
					log.Error(err)
					continue
				}
			}
		}
	}

	if len(needFeeAddress) > 0 {
		//至少要大于两笔手续费，因为自身打手续费也要一笔
		feeAddr, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 3, "coin_type": "btc-stx", "app_id": mchId}.
			And(builder.Expr("amount >= ? and forzen_amount = 0", c.cfg.NeedFee*2)), 1)
		if err != nil || len(feeAddr) == 0 {
			return fmt.Errorf("查找手续费地址错误： %v,手续费地址数量= %d", err, len(feeAddr))
		}
		feeAddress := feeAddr[0]

		amt, err := decimal.NewFromString(feeAddress.Amount)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		if amt.Cmp(decimal.NewFromFloat(c.cfg.AlarmFee)) < 0 {
			ErrDingBot.NotifyStr(fmt.Sprintf("%s 手续费不足报警,当前手续费=%v", c.coinName, amt))
			log.Errorf("%s 手续费不足报警,当前手续费=%v", c.coinName, amt)
		}

		numsDec := amt.Div(decimal.NewFromFloat(c.cfg.NeedFee))
		nums := int(numsDec.Floor().IntPart() - 1)
		log.Infof("可以打手续地址数量： %d", nums)
		if nums > len(needFeeAddress) {
			nums = len(needFeeAddress)
		}
		//生成手续费订单
		feeApply := &entity.FcTransfersApply{
			Username:   "Robot",
			CoinName:   c.coinName,
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
			Address:     feeAddress.Address,
			AddressFlag: "from",
			Status:      0,
		})
		orderReq := &transfer.StxOrderRequest{}

		for i := 0; i < nums; i++ {
			applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
				Address:     needFeeAddress[i],
				AddressFlag: "to",
				Status:      0,
			})
			orderAddress := transfer.StxToAddrAmount{
				ToAddr:   needFeeAddress[i],
				ToAmount: decimal.NewFromFloat(c.cfg.NeedFee).Shift(int32(8)).IntPart(),
			}
			orderReq.ToAddrs = append(orderReq.ToAddrs, orderAddress)
		}
		appId, err := feeApply.TransactionAdd(applyAddresses)
		if err == nil {
			//填充参数
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = feeApply.OutOrderid
			orderReq.OrderNo = feeApply.OrderId
			orderReq.MchName = mchName
			orderReq.CoinName = "BTC"

			orderReq.FromAddress = feeAddress.Address
			orderReq.ChanegeAddress = feeAddress.Address //	找零地址就是手续费地址
			orderReq.TransferFee = 0
			txid, err := c.walletServerCreateHot(orderReq)
			if err != nil {
				log.Errorf("打手续费失败： ，Err=【%v】", err)
				return err
			}
			log.Infof("%s打手续费成功，txid=【%s】", c.coinName, txid)
		}
	}
	return nil
}

func (c *CollectStxJob) walletServerCreateHot(orderReq *transfer.StxOrderRequest) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/blockstack/Transfer", c.cfg.Url), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%d],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddrs[0].ToAddr, orderReq.ToAddrs[0].ToAmount, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeStxTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result["data"].(string), nil
}

func (c *CollectStxJob) needTransferFee(address string, minAmount float64) (bool, decimal.Decimal) {
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(
		builder.Expr("type =? and address=? and coin_type=? and amount>=?", 2, address, c.feeName, c.cfg.NeedFee), 1)
	if err != nil {
		log.Errorf("get fee address balance error,err=%v", err)
		return false, decimal.NewFromFloat(0)
	}
	if fromAddrs != nil && len(fromAddrs) > 0 {
		amount, err := decimal.NewFromString(fromAddrs[0].Amount)
		if err != nil {
			return false, decimal.NewFromFloat(0)
		}
		return false, amount
	}
	return true, decimal.NewFromFloat(0)
}
