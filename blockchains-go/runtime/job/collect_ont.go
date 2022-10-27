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

type CollectOntJob struct {
	coinName string
	feeName  string
	cfg      conf.Collect2
}

const (
	OngDecimal int32 = 9
)

func NewCollectOntJob(cfg conf.Collect2) cron.Job {
	//initDingErrBot()
	return CollectOntJob{
		coinName: "ont",
		feeName:  "ong",
		cfg:      cfg,
	}
}

func (c CollectOntJob) Run() {
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

func (c *CollectOntJob) collect(mchId int, mchName string) error {
	//先获取Ont地址
	OntCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(OntCoins) == 0 {
		return errors.New("do not find Ont coin")
	}
	//获取所有合约地址
	pid := OntCoins[0].Id
	AllCoins, err1 := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err1 != nil {
		return err1
	}
	//把ong放在列表最后，因为如果ong在前面，会导致后面的币种没有手续费而无法归集，会多打一次手续费
	var (
		coins   []*entity.FcCoinSet
		ongCoin *entity.FcCoinSet
	)
	coins = append(coins, OntCoins...)
	for _, coin := range AllCoins {
		if strings.ToLower(coin.Name) == c.feeName {
			ongCoin = coin
		} else {
			coins = append(coins, coin)
		}
	}
	coins = append(coins, ongCoin)
	log.Infof("执行币种数量:%d", len(coins))
	if len(coins) > 0 {
		//手续费交易
		pendingFeeTx := make(map[string]bool)
		//已完成归集地址
		completeAddress := make(map[string]decimal.Decimal)
		//设一个默认值
		completeAddress["hoo"] = decimal.NewFromFloat(0)

		for _, coin := range coins {
			toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
				"type":        address.AddressTypeCold,
				"status":      address.AddressStatusAlloc,
				"platform_id": mchId,
				"coin_name":   c.coinName,
			})
			if err != nil || len(toAddrs) == 0 {
				log.Errorf("%s don't have cold address", mchName)
				continue
			}

			if mchId == 1 {
				if toAddrs[0] != "AHHNKyeYyVrP15j1Xwd2XRmadR6AhSsmyt" {
					log.Errorf("%s address error", mchName)
					return nil
				}
			}

			thresh := 1.0
			if coin.Name == c.coinName {
				thresh = c.cfg.MinAmount
			}
			//获取有余额的地址
			fAA, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >=  ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			if err != nil {
				log.Errorf("查询归集数据异常:%s", err.Error())
				continue
			}
			if len(fAA) == 0 {
				//log.Errorf("%s don't have need collected address", mchName)
				continue
			}
			collectAddrs := make([]*entity.FcAddressAmount, 0)
			for _, fromAddr := range fAA {
				//需要判断是否有足够的手续费
				//判断代币转账是否需要转手续费
				isNeed, amount := c.needTransferFee(fromAddr.Address, c.cfg.NeedFee)
				if isNeed {
					//加入到转手续费中
					//加入到转手续费中
					if _, ok := pendingFeeTx[fromAddr.Address]; !ok {
						pendingFeeTx[fromAddr.Address] = true
					}
				} else {
					//把amount 转换精度
					amount = amount.Shift(OngDecimal)

					//如果这个地址上一个代币归集过一笔交易
					if _, ok := completeAddress[fromAddr.Address]; ok {
						v := completeAddress[fromAddr.Address]
						tmpAmount := v.Sub(decimal.NewFromInt(int64(200000 * 500)))
						//判断手续费是否还足够转账，如果足够，就继续转账
						if tmpAmount.GreaterThanOrEqual(decimal.NewFromInt(int64(200000 * 500))) {
							collectAddrs = append(collectAddrs, fromAddr)
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
					//		tmpAmount := v.Sub(decimal.NewFromInt(int64(20000 * 500)))
					//		//判断手续费是否还足够转账，如果足够，就继续转账
					//		if tmpAmount.GreaterThanOrEqual(decimal.NewFromInt(int64(20000 * 500))) {
					//			collectAddrs = append(collectAddrs, fromAddr)
					//		}
					//	} else {
					//		//如果上一笔代币没有转过这个地址，把它记录到缓存中
					//		completeAddress[fromAddr.Address] = amount
					//		//添加这笔归集
					//		log.Infof("当前执行币种：%s,商户：%d", coin.Name, mchId)
					//		log.Infof("添加归集:%s,币种：%s,金额：%s", fromAddr.Address, fromAddr.CoinType, fromAddr.Amount)
					//
					//		collectAddrs = append(collectAddrs, fromAddr)
					//	}
					//}
				}

			}

			if len(collectAddrs) > 0 {
				//处理需要归集的地址
				for _, from := range collectAddrs {

					log.Infof("执行币种地址:%s,币种：%s,金额：%s", from.Address, coin.Name, from.Amount)
					//生成归集订单
					cltApply := &entity.FcTransfersApply{
						Username:   "Robot",
						Department: "blockchains-go",
						Applicant:  mchName,
						OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
						OrderId:    util.GetUUID(),
						Operator:   "Robot",
						CoinName:   coin.Name,
						Type:       "gj",
						Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
						Lastmodify: util.GetChinaTimeNow(),
						AppId:      mchId,
						Source:     1,
						Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
						Createtime: time.Now().Unix(),
					}

					amountFrom, _ := decimal.NewFromString(from.Amount)
					amount := amountFrom.Shift(int32(coin.Decimal))
					//如果是ong归集，减去手续费
					if strings.ToLower(coin.Name) == c.feeName {
						realFee := int64(100000 * 500)
						fee := decimal.NewFromInt(realFee)
						amount = amount.Sub(fee)
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
						orderReq := &transfer.OntOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = coin.Name

						//orderReq.CoinType = coin.Name
						orderReq.FromAddress = from.Address
						orderReq.ToAddress = to
						orderReq.Amount = amount.IntPart()
						orderReq.GasLimit = 0
						orderReq.GasPrice = 0
						txid, errTX := c.walletServerTransferHot(orderReq)
						if errTX != nil {
							log.Errorf("归集错误，err : %v", errTX)
						} else {
							log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s],txid=[%s]",
								strings.ToUpper(coin.Name), mchId, appId, from.Address, to, amount.String(), txid)
						}
					}
				}
			}

			//处理所有需要打手续费的地址
			if len(pendingFeeTx) > 0 {
				//查找手续费地址
				feeAddress := &entity.FcAddressAmount{}
				has, err := feeAddress.Get(builder.Eq{
					"app_id":    mchId,
					"coin_type": c.feeName,
					"type":      3,
				})
				if err == nil && has {
					amt, err := decimal.NewFromString(feeAddress.Amount)
					if err != nil {
						return fmt.Errorf("%v", err)
					}
					if amt.Cmp(decimal.NewFromFloat(c.cfg.AlarmFee)) < 0 {
						if ErrDingBot == nil {
							//InitDingErrBot()
						}
						//ErrDingBot.NotifyStr(fmt.Sprintf("ONG 手续费不足报警,当前手续费=%s", amt.String()))
						log.Errorf("ONG 手续费不足报警,当前手续费=%s", amt.String())
					}
					for feeAddr := range pendingFeeTx {
						log.Infof("执行币种地址222:%s", feeAddr)
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
							Purpose:    "自动打手续费",
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
							orderReq := &transfer.OntOrderRequest{}
							orderReq.ApplyId = appId
							orderReq.OuterOrderNo = feeApply.OutOrderid
							orderReq.OrderNo = feeApply.OrderId
							orderReq.MchId = int64(mchId)
							orderReq.MchName = mchName
							orderReq.CoinName = c.feeName

							//orderReq.CoinType = c.feeName
							orderReq.FromAddress = feeAddress.Address
							orderReq.ToAddress = feeAddr
							orderReq.Amount = decimal.NewFromFloat(c.cfg.NeedFee).Shift(OngDecimal).IntPart()
							orderReq.GasLimit = 0
							orderReq.GasPrice = 0
							txid, errTX := c.walletServerTransferHot(orderReq)
							if errTX != nil {
								log.Errorf("打ong手续费，err : %v", errTX)
							} else {
								log.Infof("成功打ong手续费%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s],txid=[%s]",
									strings.ToUpper(coin.Name), mchId, appId, feeAddress.Address, feeAddr, amt.String(), txid)
							}
						}
					}
				} else {
					log.Errorf("查找手续费地址错误,Err=[%v],Has=[%v]", err, has)
				}

			}

		}
	}
	return nil
}

func (c *CollectOntJob) walletServerTransferHot(orderReq *transfer.OntOrderRequest) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/Transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("post data to service error,Err=[%v]", err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", orderReq.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", orderReq.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s：data:%s", orderReq.OuterOrderNo, string(data))
	}
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，data:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.(string), nil
}

func (c *CollectOntJob) needTransferFee(address string, minAmount float64) (bool, decimal.Decimal) {
	//获取有余额的地址
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(
		builder.Expr("type =? and address=? and coin_type=? and amount>=?", 2, address, c.feeName, minAmount), 1)
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
