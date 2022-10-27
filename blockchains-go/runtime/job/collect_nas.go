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

type CollectNasJob struct {
	coinName string
	cfg      conf.Collect2
	limitMap sync.Map
}

func NewCollectNasJob(cfg conf.Collect2) cron.Job {
	return CollectNasJob{
		coinName: "nas",
		cfg:      cfg,
		limitMap: sync.Map{}, //初始化限制表
	}
}

var NasCoinId = 0

func (c CollectNasJob) Run() {
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

func (c *CollectNasJob) collect(mchId int, mchName string) error {
	//获取所有nas币种信息
	var (
		coins []*entity.FcCoinSet
		err   error
	)
	//先获取地址
	nasCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(nasCoins) == 0 {
		return errors.New("do not find nas coin")
	}
	//获取所有合约地址
	pid := nasCoins[0].Id
	NasCoinId = nasCoins[0].Id
	coins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err != nil {
		return err
	}
	coins = append(coins, nasCoins...)
	if len(coins) > 0 {
		fee := decimal.NewFromFloat(c.cfg.NeedFee).Mul(decimal.NewFromInt(200000))
		pendingFeeTx := make(map[string]int)
		var completeAddress []interface{}
		for _, coin := range coins {
			//查看是否需要归集
			if coin.IsCollect == 0 {
				log.Infof("代币 %s 未开启归集", coin.Name)
				continue
			}
			log.Infof("代币：%s,归集开启", coin.Name)

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

			//获取有余额的地址
			fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			if err != nil {
				continue
			}
			if len(fromAddrs) == 0 {
				continue
			}
			log.Infof(" %s 查询到满足归集条件的地址个数： %d", coin.Name, len(fromAddrs))
			collectAddrs := make([]*entity.FcAddressAmount, 0)
			if strings.ToLower(coin.Name) != c.coinName {
				//如果是代币归集，那么我们还需要考虑是否足够的nas手续费
				//过滤出来需要打手续费的地址
				needFee := decimal.NewFromFloat(c.cfg.NeedFee).Mul(decimal.NewFromInt(200000))
				needFee = needFee.Shift(-18)
				nf, _ := needFee.Float64()
				for _, fromAddr := range fromAddrs {
					isNeed, _ := c.needTransferFee(fromAddr.Address, nf)
					if isNeed {
						if _, ok := pendingFeeTx[fromAddr.Address]; !ok {
							pendingFeeTx[fromAddr.Address] += 1
						}
					} else {
						collectAddrs = append(collectAddrs, fromAddr)
					}
				}
			} else {
				//nas 归集
				collectAddrs = fromAddrs
			}
			if len(collectAddrs) > 0 {

				//生成归集订单
				var (
					totalNum   int // 需要归集的总数量
					successNum int // 完成归集的数量
					lockNum    int // 被锁定地址的数量
					failNum    int // 归集失败的数量
				)
				totalNum = len(collectAddrs)
				for _, from := range collectAddrs {
					//这个地址已经归集了一笔交易，不再归集
					if inArray(from.Address, completeAddress) {
						lockNum++
						continue
					}
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
					to := toAddrs[rand.Intn(len(toAddrs))] //随机地址

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
						orderReq := &transfer.NasOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.ApplyCoinId = int64(coin.Id)
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName

						orderReq.FromAddress = from.Address
						orderReq.ToAddress = to
						amount, _ := decimal.NewFromString(from.Amount)

						if coin.Name != c.coinName { //如果是代币归集
							orderReq.Token = coin.Token
							orderReq.Amount = amount.Shift(int32(coin.Decimal)).String()
						} else {
							//主链归集，要剪掉手续费
							orderReq.Amount = amount.Shift(int32(coin.Decimal)).Sub(fee).String()
						}
						txid, err := c.walletServerCollectHot(orderReq)
						if err != nil {
							failNum++
							log.Errorf("%s 归集失败： %v", coin.Name, err)
						} else {
							successNum++
							log.Infof("%s 归集成功： %s", coin.Name, txid)
							//归集成功后，添加到已完成数组里面
							completeAddress = append(completeAddress, from.Address)
						}
					} else {
						failNum++
						log.Errorf("生成applyId error，Err=[%v]", err)
					}
				}
				log.Infof("%s 归集完成，总共需要归集数量： %d，成功归集数量： %d， 失败归集数量： %d，锁定地址数量： %d",
					coin.Name,
					totalNum,
					successNum,
					failNum,
					lockNum)
			}
		}
		//如果需要打手续费
		if len(pendingFeeTx) > 0 {
			var (
				totalNum   int
				successNum int
				lockNum    int
				failNum    int
			)
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
				amt, err := decimal.NewFromString(feeAddress.Amount)
				if err != nil {
					return fmt.Errorf("%v", err)
				}
				if amt.LessThan(decimal.NewFromFloat(c.cfg.AlarmFee)) {
					if mchName == "hoo" {
						ErrDingBot.NotifyStr(fmt.Sprintf("商户:%s\n手续费地址:%s\n当前手续费:%s\n手续费报警阈值:%f",
							mchName, feeAddress.Address, feeAddress.Amount, c.cfg.AlarmFee))
					}
					return fmt.Errorf("alarm fee %v", amt)
				}
				totalNum = len(pendingFeeTx)
				log.Infof("查询到需要打手续费的地址个数为： %d", totalNum)
				for feeAddr, num := range pendingFeeTx {
					if successNum >= 5 {
						break
					}
					if inArray(feeAddr, completeAddress) {
						lockNum++
						continue
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
					if c.isUnFreezeAddress(feeAddr) {
						applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
							Address:     feeAddr,
							AddressFlag: "to",
							Status:      0,
						})

						appId, err := feeApply.TransactionAdd(applyAddresses)
						if err == nil {
							//开始请求钱包服务归集
							orderReq := &transfer.NasOrderRequest{}
							orderReq.ApplyId = appId
							orderReq.ApplyCoinId = int64(NasCoinId)
							orderReq.OuterOrderNo = feeApply.OutOrderid
							orderReq.OrderNo = feeApply.OrderId
							orderReq.MchId = int64(mchId)
							orderReq.MchName = mchName
							orderReq.CoinName = c.coinName

							orderReq.FromAddress = feeAddress.Address
							orderReq.ToAddress = feeAddr

							amount := decimal.NewFromFloat(c.cfg.NeedFee * float64(num)).Mul(decimal.NewFromInt(200000))
							orderReq.Amount = amount.String() //nas -> wei
							txid, err := c.walletServerCollectHot(orderReq)
							if err != nil {
								failNum++
								log.Errorf("Nas打手续费error,Err=[%v]", err)
							}
							successNum++
							log.Infof("Nas打手续费success,Txid=[%s]", txid)
							time.Sleep(time.Second * 30)
						} else {
							failNum++
						}
					}
				}
				log.Infof("Nas打手续费完成，总共需要打手续费：%d，成功打手续费： %d，失败打手续费： %d，锁定地址数量： %d",
					totalNum, successNum, failNum, lockNum)
			}
		}
		//wg.Wait()
	}
	return nil
}

//创建交易接口参数
func (c *CollectNasJob) walletServerCollectHot(orderReq *transfer.NasOrderRequest) (string, error) {

	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/nebulas/Transfer", c.cfg.Url), c.cfg.User, c.cfg.Password, orderReq)
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

func (c *CollectNasJob) needTransferFee(address string, minAmount float64) (bool, decimal.Decimal) {
	// 获取 nas 余额是否足够一笔归集
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
}

func (c *CollectNasJob) isUnFreezeAddress(address string) bool {
	//不做冻结处理
	return true
	//v2
	isContainAddress := false
	isUnFreeze := false
	c.limitMap.Range(func(key, value interface{}) bool {
		lastTxTime := value.(int64)
		//添加16分钟冻结时间
		freezeTime := time.Unix(lastTxTime, 0).Add(time.Minute * time.Duration(20)).Unix()
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

func inArray(need interface{}, needArr []interface{}) bool {
	if needArr == nil || len(needArr) == 0 {
		return false
	}
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}
