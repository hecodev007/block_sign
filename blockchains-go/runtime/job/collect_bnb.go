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
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

//var ErrDingBot *dingding.DingBot

type CollectBNBJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectBNBJob(cfg conf.Collect2) cron.Job {
	////钉钉通知
	//ErrDingBot = &dingding.DingBot{
	//	Name:   "ding-robot-merge-bnb",
	//	Token:  "e73c7441c796143b2c374b0f5a87efc59bdd37c950805c831d5cd46014e9814d",
	//	Source: make(chan []byte, 50),
	//	Quit:   make(chan struct{}),
	//}
	//ErrDingBot.Start()
	//initDingErrBot()
	return CollectBNBJob{
		coinName: "bnb",
		cfg:      cfg,
	}
}

func (c CollectBNBJob) Run() {
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

//其实建议维护单一币种
func (c *CollectBNBJob) collect(mchId int, mchName string) error {
	//先获取bnb地址
	bnbCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(bnbCoins) == 0 {
		return errors.New("do not find bnb coin")
	}
	//获取所有合约地址
	pid := bnbCoins[0].Id
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err != nil {
		return err
	}
	coins = append(coins, bnbCoins...)
	//手续费交易
	pendingFeeTx := make(map[string]bool)

	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	if len(coins) > 0 {
		completeAddress := make(map[string]decimal.Decimal)
		completeAddress["hoo"] = decimal.NewFromFloat(0)
		for _, coin := range coins {
			log.Infof("商户：%s,执行币种：%s", mchName, coin.Name)
			toAddrs := make([]string, 0)
			if mchId == 1 {
				toAddrs = append(toAddrs, "bnb1u8cg55dw6ls8z5ht3sezpu4jm09lu6t9dm34qy")
				toAddrs = append(toAddrs, "bnb1x2azunpdmd5ywd6xpe0ne0rn7stesh3kp8thet")
				toAddrs = append(toAddrs, "bnb1vcdqay0lnt87eh8a89gl3ngk525au5y7n7m98m")
			} else {
				toAddrs, err = entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
					"type":        address.AddressTypeCold,
					"status":      address.AddressStatusAlloc,
					"platform_id": mchId,
					"coin_name":   c.coinName,
				})
				if err != nil || len(toAddrs) == 0 {
					//fmt.Errorf("%s don't hava cold address", mchName)
					continue
				}
			}

			//thresh := 0.9
			//if coin.Name == c.coinName {
			//	thresh = c.cfg.MinAmount
			//}

			thresh := 1.0
			if strings.ToLower(coin.Name) == c.coinName {
				thresh = c.cfg.MinAmount
			}

			//数据库设置金额
			collectThreshold, _ := decimal.NewFromString(coin.CollectThreshold)
			collectThresholdFloat, _ := collectThreshold.Float64()
			if collectThresholdFloat <= 0 {
				log.Infof("代币：%s,没有设置参数，使用默认金额：%v", coin.Name, thresh)
			} else {
				thresh = collectThresholdFloat
			}
			log.Infof("%s 归集的最小金额为： %f", coin.Name, thresh)

			//获取有余额的地址
			fAA, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >=  ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			if err != nil {
				log.Errorf("查询归集数据异常:%s", err.Error())
				continue
			}
			if len(fAA) == 0 {
				//log.Errorf("%s don't hava need collected address", mchName)
				continue
			}

			//if coin.Name != c.coinName {
			//	//如果是代币归集，那么我们还需要考虑是否足够的bnb手续费
			//	//过滤出来需要打手续费的地址
			//	for _, fromAddr := range fAA {
			//		if c.needTransferFee(fromAddr.Address, c.cfg.NeedFee) {
			//			if _, ok := pendingFeeTx[fromAddr.Address]; !ok {
			//				pendingFeeTx[fromAddr.Address] = true
			//			}
			//		} else {
			//			collectAddrs = append(collectAddrs, fromAddr)
			//		}
			//	}
			//} else {
			//	collectAddrs = fAA
			//}
			collectAddrs := make([]*entity.FcAddressAmount, 0)
			for _, fromAddr := range fAA {
				//判断如果是代币
				if coin.Name != c.coinName {
					//判断代币转账是否需要转手续费
					isNeed, amount := c.needTransferFee(fromAddr.Address, c.cfg.NeedFee)
					if isNeed {
						//加入到转手续费中
						if _, ok := pendingFeeTx[fromAddr.Address]; !ok {
							pendingFeeTx[fromAddr.Address] = true
						}
					} else {
						if _, ok := completeAddress[fromAddr.Address]; ok {
							v := completeAddress[fromAddr.Address]
							//减掉上一个代币归集所用掉的手续费
							tmpAmount := v.Sub(decimal.NewFromFloat(0.000375))
							//判断手续费是否还足够转账，如果足够，就继续转账
							if tmpAmount.GreaterThanOrEqual(decimal.NewFromFloat(0.000375)) {
								collectAddrs = append(collectAddrs, fromAddr)
								//todo 如果要维护临时金额，那么需要更新
								completeAddress[fromAddr.Address] = tmpAmount
							}
						} else {
							//如果上一笔代币没有转过这个地址，把它记录到缓存中
							completeAddress[fromAddr.Address] = amount
							//添加这笔归集
							collectAddrs = append(collectAddrs, fromAddr)
						}
						//for k, v := range completeAddress {
						//	//如果这个地址上一个代币归集过一笔交易
						//	if k == fromAddr.Address {
						//		//减掉上一个代币归集所用掉的手续费
						//		tmpAmount := v.Sub(decimal.NewFromFloat(0.000375))
						//		//判断手续费是否还足够转账，如果足够，就继续转账
						//		if tmpAmount.GreaterThanOrEqual(decimal.NewFromFloat(0.000375)) {
						//			collectAddrs = append(collectAddrs, fromAddr)
						//		}
						//	} else {
						//		//如果上一笔代币没有转过这个地址，把它记录到缓存中
						//		completeAddress[fromAddr.Address] = amount
						//		//添加这笔归集
						//		collectAddrs = append(collectAddrs, fromAddr)
						//	}
						//}
					}
				} else {
					//如果是bnb归集
					if _, ok := completeAddress[fromAddr.Address]; !ok {
						//如果没归集过代币，那么归集bnb
						collectAddrs = append(collectAddrs, fromAddr)
					}
				}
			}

			if len(collectAddrs) > 0 {
				//判断该地址是否已经处理过
				//write by jun 2020/5/7
				for _, from := range collectAddrs {

					//生成归集订单
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
					if coin.Name == c.coinName {
						//签名端精度计算不对，暂时冗余0.000001
						amount = amount.Sub(fee).Sub(decimal.NewFromFloat(0.00001))
					}
					applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
					to := toAddrs[rand.Intn(len(toAddrs))] //随机取个地址
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
						orderReq := &transfer.BNBOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName
						orderReq.FromAddress = from.Address
						orderReq.ToAddress = to
						orderReq.Token = strings.ToUpper(coin.Name)
						orderReq.Quantity = amount.Shift(int32(coin.Decimal)).String()
						if err := c.walletServerCreateCold(orderReq); err != nil {
							log.Errorf("err : %v", err)
						} else {
							log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s]",
								strings.ToUpper(coin.Name), mchId, appId, from.Address, to, amount.String())
						}
					}
				}
			}
		}

		if len(pendingFeeTx) > 0 {

			//查找手续费地址
			feeAddress := &entity.FcAddressAmount{}
			has, err := feeAddress.Get(builder.In("address",
				builder.Select("address").From("fc_generate_address_list").
					Where(builder.Eq{
						"type":        3,
						"status":      2,
						"platform_id": mchId,
						"coin_name":   c.coinName,
					})).And(builder.Eq{
				"app_id":    mchId,
				"coin_type": c.coinName,
				"type":      3,
			}))

			if err == nil && has {
				amt, err := decimal.NewFromString(feeAddress.Amount)
				if err != nil {

					return fmt.Errorf("%v", err)
				}
				if amt.Cmp(decimal.NewFromFloat(c.cfg.AlarmFee)) < 0 {
					ErrDingBot.NotifyStr(fmt.Sprintf("BNB 手续费不足报警,当前手续费=%v", amt))
					log.Errorf("BNB 手续费不足报警,当前手续费=%v", amt)
				}
				maxCount := len(pendingFeeTx)
				if maxCount > c.cfg.MaxCount {
					maxCount = c.cfg.MaxCount
				}
				for feeAddr := range pendingFeeTx {
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

					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
						Address:     feeAddr,
						AddressFlag: "to",
						Status:      0,
					})
					appId, err := feeApply.TransactionAdd(applyAddresses)
					if err == nil {
						//开始请求钱包服务归集
						orderReq := &transfer.BNBOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = feeApply.OutOrderid
						orderReq.OrderNo = feeApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName

						orderReq.FromAddress = feeAddress.Address
						orderReq.ToAddress = feeAddr
						orderReq.Token = strings.ToUpper(c.coinName)
						orderReq.Quantity = decimal.NewFromFloat(c.cfg.NeedFee).Shift(8).String() //bnb
						if err := c.walletServerCreateCold(orderReq); err != nil {
							log.Errorf("err : %v", err)
						} else {
							log.Infof("成功打BNB手续费,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s]",
								mchId, appId, feeAddress.Address, feeAddr, decimal.NewFromFloat(c.cfg.NeedFee).String())
						}
					}
					maxCount--
					if maxCount <= 0 {
						break
					}
				}
			}
		}

	}
	return nil
}

func (c *CollectBNBJob) walletServerCreateCold(orderReq *transfer.BNBOrderRequest) error {
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Quantity, err)
	}
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

////创建交易接口参数
//func (c *CollectBNBJob) walletServerFee(orderReq *transfer.BNBTransferFeeReq) error {
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
//	if err != nil {
//		return err
//	}
//	dd, _ := json.Marshal(orderReq)
//	log.Infof("%s fee send :%s", c.coinName, string(dd))
//	log.Infof("%s fee resp :%s", c.coinName, string(data))
//	result := transfer.DecodeWalletServerRespOrder(data)
//	if result == nil {
//		return fmt.Errorf("walletServerFee 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
//	}
//	if result.Code != 0 {
//		log.Error(result)
//		return fmt.Errorf("walletServerFee 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
//	}
//	return nil
//}

func (c *CollectBNBJob) needTransferFee(address string, minAmount float64) (bool, decimal.Decimal) {

	//modify : 修改成从数据库拿取bnb余额 write by jun --> 2020/5/6
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(
		builder.Expr("type =? and address=? and coin_type=? and amount>=?", 2, address, c.coinName, minAmount), 1)
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
	//data, err := util.Get(fmt.Sprintf("http://bnb.rylink.io:30080/api/v1/account/%s", addresses[0]))
	//if err != nil {
	//	return false
	//}
	//if data == nil {
	//	return false
	//}
	//result, err := transfer.DecodeBNBBalanceResp(data)
	//if err != nil {
	//	return false
	//}
	//if len(result.Balances) <= 0 {
	//	return false
	//}
	//var (
	//	bnbAmount decimal.Decimal
	//	errA      error
	//)
	//for _, b := range result.Balances {
	//	if b.Symbol == "BNB" {
	//		bnbAmount, errA = decimal.NewFromString(b.Free)
	//		if errA != nil {
	//			return false
	//		}
	//	}
	//}
	//minAmt := decimal.NewFromFloat(minAmount)
	//if bnbAmount.Cmp(minAmt) < 0 {
	//	return true
	//}
	//return false
}
