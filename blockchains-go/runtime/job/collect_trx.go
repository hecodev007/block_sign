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
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectTrxJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectTrxJob(cfg conf.Collect2) cron.Job {
	return CollectTrxJob{
		coinName: "trx",
		cfg:      cfg,
	}
}

func (c CollectTrxJob) Run() {
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

func (c *CollectTrxJob) walletServerCreateHot(orderReq *transfer.TrxOrderRequest) (string, error) {
	url := fmt.Sprintf("%s/v1/%s/transfer", c.cfg.Url, c.coinName)
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	data, err := util.PostJsonByAuth(url, c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}

	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if err != nil || result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,err:%v", orderReq.OuterOrderNo, err)
	}

	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result.Data.(string), nil
}

func (c *CollectTrxJob) collect(mchId int, mchName string) error {
	// 1. 先查找代代币的coin配置
	trxCoin, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(trxCoin) == 0 {
		return fmt.Errorf("do not find %s coin", c.coinName)
	}
	pid := trxCoin[0].Id
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err != nil {
		return err
	}
	coins = append(coins, trxCoin...)

	if len(coins) > 0 {
		var (
			pendingFeeTx     = make(map[string]decimal.Decimal)
			fee              = decimal.NewFromFloat(c.cfg.NeedFee)
			compeleteAddress = make(map[string]bool)
		)
		//2. 获取归集的目标冷地址
		toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
			"type":        address.AddressTypeCold,
			"status":      address.AddressStatusAlloc,
			"platform_id": mchId,
			"coin_name":   c.coinName,
		})
		if err != nil || len(toAddrs) == 0 {
			return fmt.Errorf("%s don't have cold address,err: %v", mchName, err)
		}
		to := toAddrs[0]
		if mchId == 1 && to != "TYpUnxPMiif8RyqUmvEKNpSVL2Fv76ZhVq" {
			return fmt.Errorf("%s err address,err: %v", mchName, err)
		}

		for _, coin := range coins {
			//3. 查看是否需要归集
			if coin.IsCollect == 0 {
				log.Infof("代币 %s 未开启归集", coin.Name)
				continue
			}
			log.Infof("代币：%s,归集开启", coin.Name)
			thresh := 1.0
			if strings.ToLower(coin.Name) == c.coinName {
				thresh = c.cfg.MinAmount
			}
			//4. 数据库设置金额
			collectThreshold, _ := decimal.NewFromString(coin.CollectThreshold)
			collectThresholdFloat, _ := collectThreshold.Float64()
			if collectThresholdFloat <= 0 {
				log.Infof("代币：%s,没有设置参数，使用默认金额：%v", coin.Name, thresh)
			} else {
				thresh = collectThresholdFloat
			}
			//5. 获取有余额的地址
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
			// 6. 判断代币是否需要打手续费
			if strings.ToLower(coin.Name) != c.coinName {
				for _, fromAddr := range fromAddrs {
					if c.needTransferFee(fromAddr.Address, c.cfg.NeedFee) {
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
			log.Infof(" %s 查询到满足归集条件的地址个数： %d,大手续费地址个数：%d", coin.Name, len(collectAddrs), len(fromAddrs)-len(collectAddrs))
			//log.Infof("%s 查询到需要归集数目为：%d",coin.Name,len(collectAddrs))
			if len(collectAddrs) > 0 {
				var (
					totalNum, successNum, failNum, freeznNum int
				)
				totalNum = len(collectAddrs)
				for _, from := range collectAddrs {
					if strings.ToLower(coin.Name) == c.coinName && compeleteAddress[from.Address] {
						freeznNum++
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
					amount, _ := decimal.NewFromString(from.Amount)
					token := ""
					assetId := ""
					istrc10 := func(contractAddress string) bool {
						for _, a := range contractAddress {
							if a > 57 || a < 48 {
								return false
							}
						}
						return true
					}
					if coin.Name != c.coinName {
						cltApply.Eostoken = coin.Token
						cltApply.Eoskey = coin.Name
						//判断是否为trc10
						if istrc10(coin.Token) {
							assetId = coin.Token
						} else {
							//判断是否为trc20
							token = coin.Token
						}
					} else {
						//剪掉手续费
						amount = amount.Sub(fee)
						if amount.LessThanOrEqual(decimal.Zero) {
							failNum++
							log.Errorf("归集失败，归集金额(%s)小于手续费金额(%s)", amount.String(), fee.String())
							continue
						}
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
						orderReq := &transfer.TrxOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName
						orderReq.Worker = service.GetWorker(c.coinName)

						orderReq.FromAddress = from.Address
						orderReq.ToAddress = to
						orderReq.Amount = amount.Shift(int32(coin.Decimal)).String()
						orderReq.ContractAddress = token
						orderReq.FeeLimit = 10000000
						orderReq.AssetId = assetId
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
							failNum++
							orderHot.Status = int(status.BroadcastErrorStatus)
							orderHot.ErrorMsg = err.Error()
							dao.FcOrderHotInsert(orderHot)
							log.Errorf("%s归集错误,获取发送交易异常:%s", c.coinName, err.Error())
							// 写入热钱包表，创建失败
							log.Errorf(err.Error())
							continue
						}
						successNum++
						if strings.ToLower(coin.Name) != c.coinName {
							compeleteAddress[from.Address] = true
						}
						orderHot.TxId = txid
						orderHot.Status = int(status.BroadcastStatus)
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
						log.Error(err)
						failNum++
						continue
					}
				}
				log.Infof("%s 归集成功，总共需要归集数量：%d，成功数量：%d，失败数量：%d,冻结地址数量： %d", strings.ToUpper(coin.Name),
					totalNum, successNum, failNum, freeznNum)
			}
		}
		if len(pendingFeeTx) > 0 {
			log.Infof("开始打手续费")
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
				}
				var (
					totalNum   int
					successNum int
					failNum    int
				)
				totalNum = len(pendingFeeTx)
				log.Infof("查询到需要打手续费的地址个数为： %d", totalNum)
				for feeAddr, needFee := range pendingFeeTx {
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
						Address:     feeAddr,
						AddressFlag: "to",
						Status:      0,
						Lastmodify:  feeApply.Lastmodify,
					})
					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
						Address:     feeAddress.Address,
						AddressFlag: "from",
						Status:      0,
						Lastmodify:  feeApply.Lastmodify,
					})
					appId, err := feeApply.TransactionAdd(applyAddresses)
					if err == nil {
						//填充参数
						orderReq := &transfer.TrxOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = feeApply.OutOrderid
						orderReq.OrderNo = feeApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName
						orderReq.Worker = service.GetWorker(c.coinName)

						orderReq.FromAddress = feeAddress.Address
						orderReq.ToAddress = feeAddr
						orderReq.Amount = needFee.Shift(int32(trxCoin[0].Decimal)).String()

						//发送交易
						createData, _ := json.Marshal(orderReq)
						orderHot := &entity.FcOrderHot{
							ApplyId:      int(appId),
							ApplyCoinId:  trxCoin[0].Id,
							OuterOrderNo: feeApply.OutOrderid,
							OrderNo:      feeApply.OrderId,
							MchName:      mchName,
							CoinName:     c.coinName,
							FromAddress:  orderReq.FromAddress,
							ToAddress:    orderReq.ToAddress,
							Amount:       needFee.Shift(int32(trxCoin[0].Decimal)).IntPart(), //转换整型
							Quantity:     needFee.Shift(int32(trxCoin[0].Decimal)).String(),
							Decimal:      int64(trxCoin[0].Decimal),
							CreateData:   string(createData),
							Status:       int(status.UnknowErrorStatus),
							CreateAt:     time.Now().Unix(),
							UpdateAt:     time.Now().Unix(),
						}

						txid, err := c.walletServerCreateHot(orderReq)
						if err != nil {
							failNum++
							orderHot.Status = int(status.BroadcastErrorStatus)
							orderHot.ErrorMsg = err.Error()
							dao.FcOrderHotInsert(orderHot)
							log.Errorf("%s归集错误,获取发送交易异常:%s", c.coinName, err.Error())
							// 写入热钱包表，创建失败
							log.Errorf(err.Error())
							continue
						}
						successNum++
						orderHot.TxId = txid
						orderHot.Status = int(status.BroadcastStatus)
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
						log.Error(err)
						failNum++
						continue
					}
					// 配合签名程序 限制转账时间
					time.Sleep(5 * time.Second)
				}
				log.Infof("%s 打手续费成功，总共需要打手续费数量：%d，成功数量：%d，失败数量：%d", c.coinName, totalNum, successNum, failNum)
			}
		}
	}
	return nil
}

func (c *CollectTrxJob) needTransferFee(address string, minAmount float64) bool {
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(
		builder.Expr("type =? and address=? and coin_type=? and amount>=?", 2, address, c.coinName, minAmount), 1)
	if err != nil {
		log.Errorf("get fee address balance error,err=%v", err)
		return false
	}
	if fromAddrs != nil && len(fromAddrs) > 0 {
		_, err := decimal.NewFromString(fromAddrs[0].Amount)
		if err != nil {
			return false
		}
		return false
	}
	return true
}
