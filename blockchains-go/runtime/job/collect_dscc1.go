package job

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectDscc1Job struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectDscc1Job(cfg conf.Collect2) cron.Job {
	return CollectDscc1Job{
		coinName: "dscc1",
		cfg:      cfg,
	}
}

func (c CollectDscc1Job) Run() {
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

func (c *CollectDscc1Job) collect(mchId int, mchName string) error {
	////获取归集的目标冷地址
	//toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
	//	"type":        address.AddressTypeCold,
	//	"status":      address.AddressStatusAlloc,
	//	"platform_id": mchId,
	//	"coin_name":   c.coinName,
	//})
	//if len(toAddrs) == 0 {
	//	return fmt.Errorf("%s don't have cold address", mchName)
	//}
	//获取币种的配置
	dscc1Coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(dscc1Coins) == 0 {
		return fmt.Errorf("do not find %s coin", c.coinName)
	}

	pid := dscc1Coins[0].Id
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err != nil {
		return err
	}
	coins = append(coins, dscc1Coins...)
	completeAddress := make(map[string]bool)
	pendingFeeTx := make(map[string]decimal.Decimal)
	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	log.Infof("coin数量：%d", len(coins))
	for _, coin := range coins {
		//1. 查看是否需要归集
		if coin.IsCollect == 0 {
			log.Infof("代币 %s 未开启归集", coin.Name)
			continue
		}
		log.Infof("代币：%s,归集开启", coin.Name)
		toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
			"type":        address.AddressTypeCold,
			"status":      address.AddressStatusAlloc,
			"platform_id": mchId,
			"coin_name":   c.coinName,
		})

		if err != nil {
			log.Errorf("get to address error,Err=%v", err)
			continue
		}
		if len(toAddrs) == 0 {
			log.Errorf("%s don't have cold address", mchName)
			continue
		}

		var to string
		//判断该币种是否需要归集到多个地址
		if c.inArray(coin.Name) {
			log.Info("代币开启归集到多个冷地址")
			to = toAddrs[rand.Intn(len(toAddrs))]
		} else {
			to = toAddrs[0]
			//if mchId == 1 && toAddrs[0] != "0x05f0fdd0e49a5225011fff92ad85cc68e1d1f08e" {
			//	log.Infof("hoo 归集地址异常,归集冷地址为=【%s】冷地址不等于：0x05f0fdd0e49a5225011fff92ad85cc68e1d1f08e", to)
			//	continue
			//}
		}
		// 3. 设置最小的归集金额
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
		//4。 获取有余额的地址
		fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
			And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
			And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
		if err != nil {
			continue
		}
		if len(fromAddrs) == 0 {
			continue
		}
		collectAddrs := make([]*entity.FcAddressAmount, 0)
		//5. 判断代币归集是否需要手续费
		if strings.ToLower(coin.Name) != c.coinName {
			for _, fromAddr := range fromAddrs {
				isNeed, _ := c.needTransferFee(fromAddr.Address, c.cfg.NeedFee)
				if isNeed {
					if f, ok := pendingFeeTx[fromAddr.Address]; !ok {
						pendingFeeTx[fromAddr.Address] = fee
					} else {
						f = f.Add(fee)
						pendingFeeTx[fromAddr.Address] = f
					}
				} else {
					collectAddrs = append(collectAddrs, fromAddr)
				}
			}
		} else {
			collectAddrs = fromAddrs
		}
		if len(collectAddrs) > 0 {
			var (
				totalNum, lockNum, successNum, failNum int
			)
			totalNum = len(collectAddrs)
			log.Infof(" %s 查询到满足归集条件的地址个数： %d", coin.Name, totalNum)
			//获取有余额的地址
			for _, from := range collectAddrs {
				// 7. 判断是否已经完成归集，每个地址每次只归集一次
				if completeAddress[from.Address] {
					lockNum++
					continue
				}

				//判断数据库金额是否与链上金额一致
				amount, _ := decimal.NewFromString(from.Amount)
				dbAmount := amount.Shift(int32(coin.Decimal))
				chainAmountStr, err := c.getChainBalance(from.Address, coin.Token)
				if err != nil {
					log.Errorf("获取[%s]链上金额失败: %s", from.Address, err.Error())
					failNum++
					continue
				}
				chainAmount, _ := decimal.NewFromString(chainAmountStr)
				if chainAmount.LessThan(dbAmount) {
					log.Errorf("[%s]数据库金额[%s]大于链上金额[%s]，不进行归集,币种[%s]", from.Address, dbAmount, chainAmount, coin.Name)
					failNum++
					continue
				}
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
				applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
				//to := toAddrs[rand.Intn(len(toAddrs))] //随机取个地址
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
					orderReq := &transfer.Dscc1OrderRequest{}
					orderReq.ApplyId = appId
					orderReq.OuterOrderNo = cltApply.OutOrderid
					orderReq.OrderNo = cltApply.OrderId
					orderReq.MchId = int64(mchId)
					orderReq.MchName = mchName
					orderReq.CoinName = c.coinName
					orderReq.Worker = service.GetWorker(c.coinName)

					orderReq.FromAddress = from.Address
					orderReq.ToAddress = to

					//		判断是否是合约归集
					if strings.ToLower(coin.Name) != c.coinName {
						orderReq.ContractAddress = coin.Token
						orderReq.Token = coin.Name
						orderReq.Amount = amount.Shift(int32(coin.Decimal)).String()
					} else {
						//扣除手续费
						fee := decimal.NewFromFloat(c.cfg.NeedFee)
						amount = amount.Sub(fee)
						orderReq.Amount = amount.Shift(int32(coin.Decimal)).String()
					}

					//发送交易
					createData, _ := json.Marshal(orderReq)
					orderHot := &entity.FcOrderHot{
						ApplyId:      int(appId),
						ApplyCoinId:  coin.Id,
						OuterOrderNo: cltApply.OutOrderid,
						OrderNo:      cltApply.OrderId,
						MchName:      mchName,
						CoinName:     c.coinName,
						FromAddress:  orderReq.FromAddress,
						ToAddress:    orderReq.ToAddress,
						Amount:       amount.Shift(int32(coin.Decimal)).IntPart(), //转换整型
						Quantity:     amount.Shift(int32(coin.Decimal)).String(),
						Decimal:      int64(coin.Decimal),
						CreateData:   string(createData),
						Status:       int(status.UnknowErrorStatus),
						CreateAt:     time.Now().Unix(),
						UpdateAt:     time.Now().Unix(),
					}

					txid, err := c.walletServerCreateHot(orderReq)
					if err != nil {
						orderHot.Status = int(status.BroadcastErrorStatus)
						orderHot.ErrorMsg = err.Error()
						dao.FcOrderHotInsert(orderHot)
						log.Errorf("%s归集错误,获取发送交易异常:%s", c.coinName, err.Error())
						// 写入热钱包表，创建失败
						log.Errorf(err.Error())
						failNum++
						continue
					}

					orderHot.TxId = txid
					orderHot.Status = int(status.BroadcastStatus)
					successNum++
					//保存热表
					err = dao.FcOrderHotInsert(orderHot)
					if err != nil {
						err = fmt.Errorf("[%s]归集保存订单[%s]数据异常:[%s]", c.coinName, orderHot.OuterOrderNo, err.Error())
						//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
						log.Error(err.Error())
						//发送给钉钉
						// dingding.ErrTransferDingBot.NotifyStr(err.Error())
					}
				} else {
					failNum++
					log.Error(err)
					continue
				}
			}
			log.Infof("%s 归集完成，总共需要归集笔数：%d，锁定归集笔数：%d，成功归集笔数： %d，失败归集笔数：%d", strings.ToUpper(coin.Name),
				totalNum, lockNum, successNum, failNum)
		}

	}
	if len(pendingFeeTx) > 0 {
		var (
			useNum, totalNum, lockNum, successNum, failNum int
		)
		totalNum = len(pendingFeeTx)
		log.Infof("需要打手续费的笔数为：%d", totalNum)
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
				//log.Errorf("alarm fee %v", amt)
				if mchName == "hoo" {
					ErrDingBot.NotifyStr(fmt.Sprintf("商户:%s\n手续费地址:%s\n当前手续费:%s\n手续费报警阈值:%f",
						mchName, feeAddress.Address, feeAddress.Amount, c.cfg.AlarmFee))
				}
			}

			//判断手续费地址的金额是否足够
			if amt.LessThan(decimal.NewFromFloat(c.cfg.NeedFee * 2)) {
				return fmt.Errorf("%s 手续费不足，fee=%s", c.coinName, amt.String())
			}
			//判断能打几笔手续费，避免报错
			numsDec := amt.Div(decimal.NewFromFloat(c.cfg.NeedFee))
			nums, _ := numsDec.Float64()
			useNum = int(math.Ceil(nums)) - 1

			for to, feeAmount := range pendingFeeTx {
				if useNum <= 0 {
					break
				}
				if completeAddress[to] {
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
				applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
					Address:     to,
					AddressFlag: "to",
					Status:      0,
				})

				appId, err := feeApply.TransactionAdd(applyAddresses)
				if err == nil {
					//填充参数
					orderReq := &transfer.Dscc1OrderRequest{}
					orderReq.ApplyId = appId
					orderReq.OuterOrderNo = feeApply.OutOrderid
					orderReq.OrderNo = feeApply.OrderId
					orderReq.MchId = int64(mchId)
					orderReq.MchName = mchName
					orderReq.CoinName = c.coinName
					orderReq.Worker = service.GetWorker(c.coinName)

					orderReq.FromAddress = feeAddress.Address
					orderReq.ToAddress = to
					orderReq.Amount = feeAmount.Shift(int32(dscc1Coins[0].Decimal)).String()

					//发送交易
					createData, _ := json.Marshal(orderReq)
					orderHot := &entity.FcOrderHot{
						ApplyId:      int(appId),
						ApplyCoinId:  dscc1Coins[0].Id,
						OuterOrderNo: feeApply.OutOrderid,
						OrderNo:      feeApply.OrderId,
						MchName:      mchName,
						CoinName:     c.coinName,
						FromAddress:  orderReq.FromAddress,
						ToAddress:    orderReq.ToAddress,
						Amount:       feeAmount.Shift(int32(dscc1Coins[0].Decimal)).IntPart(), //转换整型
						Quantity:     feeAmount.Shift(int32(dscc1Coins[0].Decimal)).String(),
						Decimal:      int64(dscc1Coins[0].Decimal),
						CreateData:   string(createData),
						Status:       int(status.UnknowErrorStatus),
						CreateAt:     time.Now().Unix(),
						UpdateAt:     time.Now().Unix(),
					}

					txid, err := c.walletServerCreateHot(orderReq)
					if err != nil {
						orderHot.Status = int(status.BroadcastErrorStatus)
						orderHot.ErrorMsg = err.Error()
						dao.FcOrderHotInsert(orderHot)
						log.Errorf("%s归集错误,获取发送交易异常:%s", c.coinName, err.Error())
						// 写入热钱包表，创建失败
						log.Errorf(err.Error())
						failNum++
						continue
					}

					orderHot.TxId = txid
					orderHot.Status = int(status.BroadcastStatus)
					successNum++
					useNum--
					//保存热表
					err = dao.FcOrderHotInsert(orderHot)
					if err != nil {
						err = fmt.Errorf("[%s]归集保存订单[%s]数据异常:[%s]", c.coinName, orderHot.OuterOrderNo, err.Error())
						//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
						log.Error(err.Error())
						//发送给钉钉
						// dingding.ErrTransferDingBot.NotifyStr(err.Error())
					}
				} else {
					failNum++
					log.Errorf("create collect app id error,Err=%v", err)
					continue
				}
				time.Sleep(10 * time.Second)
			}
		} else {
			return fmt.Errorf(" 打手续费失败，查询fee address error,Err=%v", err)
		}
		log.Infof("打手续费成功，可以打手续费笔数：%d，总共需要打手续费笔数：%d，锁定笔数：%d，成功笔数：%d，失败笔数：%d",
			useNum, totalNum, lockNum, successNum, failNum)
	}
	return nil
}

func (c *CollectDscc1Job) walletServerCreateHot(orderReq *transfer.Dscc1OrderRequest) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeDscc1TransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result["data"].(string), nil
}

func (c *CollectDscc1Job) needTransferFee(address string, minAmount float64) (bool, decimal.Decimal) {
	// 获取 余额是否足够一笔归集
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
func (c *CollectDscc1Job) inArray(coin string) bool {
	if len(c.cfg.AssignCoins) == 0 {
		return false
	}
	for _, cc := range c.cfg.AssignCoins {
		if strings.ToLower(cc) == strings.ToLower(coin) {
			return true
		}
	}
	return false
}

func (c *CollectDscc1Job) getChainBalance(address, contract string) (string, error) {

	var req transfer.ReqGetBalanceParams
	req.Address = address
	req.CoinName = "dscc1"
	req.ContractAddress = contract
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/getBalance", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, req)
	if err != nil {
		return "", err
	}
	gbr, err := transfer.DecodeGetBalanceResp(data)
	if err != nil {
		return "", err
	}
	balance, ok := gbr.Data.(string)
	if !ok {
		return "", fmt.Errorf("%v is not string", gbr.Data)
	}
	return balance, nil
}
