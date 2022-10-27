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
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectMtrJob struct {
	coinName string
	cfg      conf.Collect2
	limitMap sync.Map
}

func NewCollectMtrJob(cfg conf.Collect2) cron.Job {
	return CollectMtrJob{
		coinName: "mtr",
		cfg:      cfg,
		limitMap: sync.Map{}, //初始化限制表
	}
}

func (c CollectMtrJob) Run() {
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

func (c *CollectMtrJob) collect(mchId int, mchName string) error {
	//start := time.Now()
	//log.Infof("=== %s collect task start ===", mchName)
	//defer log.Infof("=== %s collect task end, use time : %f s ===", mchName, time.Since(start).Seconds())
	//获取所有mtr币种信息
	var (
		coins     []*entity.FcCoinSet
		err       error
		mainCoins []*entity.FcCoinSet
	)

	mainCoins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(mainCoins) == 0 {
		return errors.New("empty main coin")
	}
	mainCoin := mainCoins[0]
	if len(c.cfg.AssignCoins) > 0 {
		coins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": mainCoin.Id}.Or(builder.Eq{"id": mainCoin.Id})).And(builder.In("name", c.cfg.AssignCoins)))
	} else {
		coins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": mainCoin.Id}.Or(builder.Eq{"id": mainCoin.Id})).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	}
	if err != nil {
		return err
	}
	if len(coins) > 0 {
		//wg := &sync.WaitGroup{}
		pendingFeeTx := make(map[string]struct{})
		useAddress := make(map[string]struct{})
		for _, coin := range coins {
			//获取归集的目标冷地址
			toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
				"type":        address.AddressTypeCold,
				"status":      address.AddressStatusAlloc,
				"platform_id": mchId,
				"coin_name":   c.coinName,
			})
			if err != nil || len(toAddrs) == 0 {
				continue
			}

			if mchId == 1 {
				for _, v := range toAddrs {
					if v != "0x6773bc5b62a9efe1efd78fe4436c681267414628" {
						if v != "0xf0c4392a3c1d29867330051a3f80729b9a756665" {
							if v != "0x8efacab8d14171bd7eb1238c1a0f5eb074a3822a" {
								log.Info("error mch1 addr")
								continue
							}
						}
					}
				}
			}

			//代币余额1.0开始
			thresh := 0.1
			if strings.ToLower(coin.Name) == c.coinName {
				thresh = c.cfg.MinAmount
			}
			//获取有余额的地址
			fromAddrsByUser, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			if err != nil {
				//log.Errorf("查询归集数据异常:%s", err.Error())
				continue
			}
			if len(fromAddrsByUser) == 0 {
				//log.Errorf("%s don't hava need collected address", mchName)
				continue
			}

			//单一币种归集
			fromAddrs := make([]*entity.FcAddressAmount, 0)
			for _, v := range fromAddrsByUser {
				if _, ok := useAddress[v.Address]; !ok {
					fromAddrs = append(fromAddrs, v)
				}
			}
			collectAddrs := make([]*entity.FcAddressAmount, 0)
			//feeAddrs := make([]string, 0)
			log.Infof("执行币种：%s", coin.Name)
			if strings.ToLower(coin.Name) != c.coinName {
				//如果是代币归集，那么我们还需要考虑是否足够的mtr手续费
				//过滤出来需要打手续费的地址
				for _, fromAddr := range fromAddrs {
					if c.needTransferFee(fromAddr.Address, c.cfg.NeedFee, thresh) {
						if _, ok := pendingFeeTx[fromAddr.Address]; !ok {
							pendingFeeTx[fromAddr.Address] = struct{}{}
						}
					} else {
						collectAddrs = append(collectAddrs, fromAddr)

					}
				}
			} else {
				//mtr 归集
				collectAddrs = fromAddrs
			}
			if len(collectAddrs) > 0 {

				//}
				for _, from := range fromAddrs {
					useAddress[from.Address] = struct{}{}
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
						Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
						Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
						Createtime: time.Now().Unix(),
						Lastmodify: util.GetChinaTimeNow(),
						Source:     1,
					}
					if coin.Name != c.coinName { //代表是代币
						cltApply.Eostoken = coin.Token
						cltApply.Eoskey = coin.Name
					}

					applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
					randNum := 4 //随机取前4个地址
					if len(toAddrs) < randNum {
						randNum = len(toAddrs)
					}
					to := toAddrs[rand.Intn(randNum)] //随机地址
					//for _, to := range toAddrs {
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
						//开始请求钱包服务归集
						toAmtFloat, _ := decimal.NewFromString(from.Amount)
						orderReq := &transfer.MtrOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName
						orderReq.FromAddr = from.Address
						orderReq.ToAddr = to

						if coin.Name != c.coinName { //如果是代币归集
							token, err := strconv.ParseInt(coin.Token, 10, 64)
							if err != nil {
								log.Infof("token转换异常:%s", coin.Token)
								continue
							}
							orderReq.Token = token
							orderReq.ToAmountInt64 = toAmtFloat.Shift(int32(coin.Decimal)).String()
						} else {
							orderReq.ToAmountInt64 = toAmtFloat.Sub(decimal.NewFromFloat(c.cfg.NeedFee)).Shift(int32(coin.Decimal)).String()
						}
						if _, err := c.walletServerCollect(orderReq); err != nil {
							log.Errorf("err : %v", err)
						}
					}
				}

			}
			//}(coin)
		}
		//如果需要打手续费
		if len(pendingFeeTx) > 0 {
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
			//查找手续费地址
			feeAddress := &entity.FcAddressAmount{}
			has, err := feeAddress.Get(builder.In("address",
				builder.Select("address").From("fc_generate_address_list").
					Where(builder.Eq{
						"type":        address.AddressTypeFee,
						"status":      2,
						"platform_id": mchId,
						"coin_name":   c.coinName,
					})).And(builder.Eq{
				"app_id":    mchId,
				"coin_type": c.coinName,
				"type":      3,
			}))
			if err == nil && has {
				//todo 商户手续费告警
				amt, err := decimal.NewFromString(feeAddress.Amount)
				if err != nil {
					//log.Errorf("%v", err)
					return fmt.Errorf("%v", err)
				}
				if amt.LessThan(decimal.NewFromFloat(c.cfg.AlarmFee)) {
					//log.Errorf("alarm fee %v", amt)
					if mchName == "hoo" {
						ErrDingBot.NotifyStr(fmt.Sprintf("商户:%s\n手续费地址:%s\n当前手续费:%s\n手续费报警阈值:%f",
							mchName, feeAddress.Address, feeAddress.Amount, c.cfg.AlarmFee))
					}
					return fmt.Errorf("alarm fee %v", amt)
				}

				feeAddrs := make([]string, 0)
				for feeAddr := range pendingFeeTx {
					useAddress[feeAddr] = struct{}{}
					applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
						Address:     feeAddress.Address,
						AddressFlag: "from",
						Status:      0,
					})
					if c.isUnFreezeAddress(feeAddr) {
						applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
							Address:     feeAddr,
							AddressFlag: "to",
							Status:      0,
						})
						feeAddrs = append(feeAddrs, feeAddr)
					}

					appId, err := feeApply.TransactionAdd(applyAddresses)
					if err == nil {
						//开始请求钱包服务归集
						orderReq := &transfer.MtrOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = feeApply.OutOrderid
						orderReq.OrderNo = feeApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName
						orderReq.FromAddr = feeAddress.Address
						orderReq.ToAddr = feeAddr
						orderReq.ToAmountInt64 = decimal.NewFromFloat(c.cfg.NeedFee).Shift(18).String() //mtr -> wei
						if _, err := c.walletServerFee(orderReq); err != nil {
							log.Errorf("err : %v", err)
						}
					}

				}
			}
		}
		//wg.Wait()
	}
	return nil
}

//创建交易接口参数
func (c *CollectMtrJob) walletServerCollect(orderReq *transfer.MtrOrderRequest) (string, error) {

	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", c.coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", c.coinName, string(data))
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.New("json unmarshal response data error")
	}
	if resp["data"] == nil {
		return "", fmt.Errorf("mtr transfer error,Err=%s", string(data))
	}
	return resp["data"].(string), nil
}

//创建交易接口参数
func (c *CollectMtrJob) walletServerFee(orderReq *transfer.MtrOrderRequest) (string, error) {

	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", c.coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", c.coinName, string(data))
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.New("json unmarshal response data error")
	}
	if resp["data"] == nil {
		return "", fmt.Errorf("mtr transfer error,Err=%s", string(data))
	}
	return resp["data"].(string), nil

}

func (c *CollectMtrJob) needTransferFee(address string, feeLimit, thresh float64) bool {
	if address == "" || feeLimit <= 0 {
		log.Info("error params")
		return false
	}
	url := fmt.Sprintf("%s/v1/%s/balance?addr=%s", c.cfg.Url, c.coinName, address)
	data, err := util.GetByAuth(url, c.cfg.User, c.cfg.Password)
	log.Infof("url:%s,data:%s", url, string(data))
	if err != nil {
		log.Infof("url:%s,error:%s,data:%s", url, err.Error(), string(data))
		return false
	}
	result := transfer.DecodeMtrBalanceResp(data)
	if result == nil {
		log.Error("result empty")
		return false
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return false
	}
	if err != nil {
		return false
	}
	minAmt := decimal.NewFromFloat(feeLimit)
	threshAmt := decimal.NewFromFloat(thresh)

	hasToken := false
	enoughFee := false

	log.Infof("address:%s", address)
	for _, v := range result.Data {
		if v.CoinName == "mtrg" {
			amountToken, _ := decimal.NewFromString(v.BalanceFloat)
			log.Infof("mtrg:%s", amountToken.String())
			if amountToken.GreaterThanOrEqual(threshAmt) {
				//需要归集，检查手续费
				hasToken = true
			}
		}
		if v.CoinName == "mtr" {
			amountToken, _ := decimal.NewFromString(v.BalanceFloat)
			log.Infof("mtr:%s", amountToken.String())
			if amountToken.GreaterThanOrEqual(minAmt) {
				//需要归集，检查手续费
				return false
			}
		}
	}
	//
	if hasToken && !enoughFee {
		//存在代币，但是不够手续费
		return true
	}

	return false
}

func (c *CollectMtrJob) isUnFreezeAddress(address string) bool {
	//v2
	isContainAddress := false
	isUnFreeze := false
	c.limitMap.Range(func(key, value interface{}) bool {
		lastTxTime := value.(int64)
		//添加16分钟冻结时间
		freezeTime := time.Unix(lastTxTime, 0).Add(time.Minute * time.Duration(16)).Unix()
		now := time.Now().Unix()
		if now >= freezeTime {
			if key.(string) == address {
				isContainAddress = true
				log.Infof("address=[%s]已解冻,可以进行出账", address)
				c.limitMap.Store(address, now) //	添加新的冻结时间
				isUnFreeze = true
			} else {
				//从map中移除
				c.limitMap.Delete(key)
			}
		} else {
			if key.(string) == address {
				isContainAddress = true
				TimeStr := func(timestamp int64) string {
					var timeLayout = "2006-01-02 T 15:04:05.000"
					return time.Unix(timestamp, 0).Format(timeLayout)
				}
				log.Infof("address=[%s]被冻结，解冻时间为：[%s],当前时间为：[%s]", address, TimeStr(freezeTime), TimeStr(now))
				isUnFreeze = false
			}
		}
		return true
	})

	if !isContainAddress {
		c.limitMap.Store(address, time.Now().Unix())
		isUnFreeze = true
	}
	return isUnFreeze
}
