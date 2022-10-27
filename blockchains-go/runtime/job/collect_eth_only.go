package job

//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/group-coldwallet/blockchains-go/conf"
//	"github.com/group-coldwallet/blockchains-go/dao"
//	"github.com/group-coldwallet/blockchains-go/entity"
//	"github.com/group-coldwallet/blockchains-go/model/address"
//	"github.com/group-coldwallet/blockchains-go/model/transfer"
//	"github.com/group-coldwallet/blockchains-go/pkg/util"
//	"github.com/group-coldwallet/blockchains-go/log"
//	"github.com/robfig/cron/v3"
//	"github.com/shopspring/decimal"
//	"math/rand"
//	"strings"
//	"sync"
//	"time"
//	"xorm.io/builder"
//)
//
//type CollectbscJob struct {
//	coinName                 string
//	cfg                      conf.Collect2
//	limitMap                 sync.Map
//	feeAddrPendingOutOrderId []string
//}
//
////只归集主链币
//func NewCollectbscJob(cfg conf.Collect2) cron.Job {
//	return CollectbscJob{
//		coinName: "eth",
//		cfg:      cfg,
//		limitMap: sync.Map{}, //初始化限制表
//	}
//}
//
//func (c CollectbscJob) Run() {
//	var (
//		mchs []*entity.FcMch
//		err  error
//	)
//	start := time.Now()
//	log.Infof("*** %s collect task start***", c.coinName)
//	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(start).Seconds())
//
//	if len(c.cfg.Mchs) != 0 {
//		mchs, err = entity.FcMch{}.Find(builder.In("platform", c.cfg.Mchs).And(builder.Eq{"status": 2}))
//	} else {
//		mchs, err = entity.FcMch{}.Find(builder.In("id", builder.Select("mch_id").From("fc_mch_service").
//			Where(builder.Eq{
//				"status":    0,
//				"coin_name": c.coinName,
//			})).And(builder.Eq{"status": 2}))
//	}
//	if err != nil {
//		log.Errorf("find platforms err %v", err)
//		return
//	}
//	wg := &sync.WaitGroup{}
//	for _, tmp := range mchs {
//		go func(mch *entity.FcMch) {
//			wg.Add(1)
//			defer wg.Done()
//			if err := c.collect(mch.Id, mch.Platform); err != nil {
//				log.Errorf(" %s ## collect err: %v", mch.Platform, err)
//			}
//		}(tmp)
//	}
//	wg.Wait()
//}
//
//func (c *CollectbscJob) collect(mchId int, mchName string) error {
//
//	mainCoin,err := dao.FcCoinSetGetByName("eth",1)
//	if err != nil {
//		return err
//	}
//	if mainCoin.IsCollect == 0 {
//		return errors.New("eth关闭了归集")
//	}
//	toAddrs:= make([]string, 0)
//	if mchId == 1 {
//		toAddrs = append(toAddrs, "0x0093e5f2a850268c0ca3093c7ea53731296487eb")
//		toAddrs = append(toAddrs, "0x002471c86e9e97d393d84bddfa7d555a7fa2917a")
//		toAddrs = append(toAddrs, "0x0055e75217ca5cb5aa8290cd966f9d85751a7993")
//	} else {
//		toAddrs, err = entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
//			"type":        address.AddressTypeCold,
//			"status":      address.AddressStatusAlloc,
//			"platform_id": mchId,
//			"coin_name":   c.coinName,
//		}.And(builder.Neq{"address": "0x9760862d09b70433a91ff27cbd069f51ef1cbd5c"}))
//		if err != nil || len(toAddrs) == 0 {
//			//fmt.Errorf("%s don't hava cold address", mchName)
//			return errors.New("获取归集地址异常")
//		}
//	}
//
//
//	//获取有余额的地址,没有代币的地址
//	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
//		And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
//		And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
//	if err != nil {
//		//log.Errorf("查询归集数据异常:%s", err.Error())
//		continue
//	}
//	if len(fromAddrs) == 0 {
//		//log.Errorf("%s don't hava need collected address", mchName)
//		continue
//	}
//
//
//
//		for _, coin := range coins {
//
//
//
//
//			//数据库设置金额
//
//
//
//			//获取有余额的地址
//			fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
//				And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
//				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
//			if err != nil {
//				//log.Errorf("查询归集数据异常:%s", err.Error())
//				continue
//			}
//			if len(fromAddrs) == 0 {
//				//log.Errorf("%s don't hava need collected address", mchName)
//				continue
//			}
//			collectAddrs := make([]string, 0)
//			//feeAddrs := make([]string, 0)
//			if strings.ToLower(coin.Name) != c.coinName {
//				//如果是代币归集，那么我们还需要考虑是否足够的eth手续费
//				//过滤出来需要打手续费的地址
//				for _, fromAddr := range fromAddrs {
//					if c.needTransferFee([]string{fromAddr}, c.cfg.NeedFee) {
//						if _, ok := pendingFeeTx[fromAddr]; !ok {
//							pendingFeeTx[fromAddr] = true
//						}
//					} else {
//						time.Sleep(500 * time.Millisecond)
//						//检查代币了链上金额是否大于0
//						tokenBalance, err := getTokenBalance(coin.Token, fromAddr)
//						if err != nil {
//							log.Errorf("检查代币金额失败：%s", err.Error())
//							continue
//						}
//						if tokenBalance.IsZero() {
//							continue
//						}
//						log.Infof("地址：%s,token:%s,精度：%d,实际金额：%s", fromAddr, coin.Name, coin.Decimal, tokenBalance.Shift(int32(coin.Decimal)).String())
//						collectAddrs = append(collectAddrs, fromAddr)
//
//					}
//				}
//			} else {
//				//eth 归集
//				collectAddrs = fromAddrs
//			}
//			if len(collectAddrs) > 0 {
//				//生成归集订单
//				cltApply := &entity.FcTransfersApply{
//					Username:   "Robot",
//					CoinName:   c.coinName,
//					Department: "blockchains-go",
//					OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
//					OrderId:    util.GetUUID(),
//					Applicant:  mchName,
//					Operator:   "Robot",
//					AppId:      mchId,
//					Type:       "gj",
//					Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
//					Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
//					Createtime: time.Now().Unix(),
//					Lastmodify: util.GetChinaTimeNow(),
//					Source:     1,
//				}
//				if coin.Name != c.coinName { //代表是代币
//					cltApply.Eostoken = coin.Token
//					cltApply.Eoskey = coin.Name
//				}
//
//				applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
//				randNum := 4 //随机取前4个地址
//				if len(toAddrs) < randNum {
//					randNum = len(toAddrs)
//				}
//				to := toAddrs[rand.Intn(randNum)] //随机地址
//
//				//减少某个地址的归集次数，因为被浏览器标记了
//				if strings.ToLower(to) == "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
//					//再次随机两次
//					for r := 0; r < 2; r++ {
//						to = toAddrs[rand.Intn(randNum)]
//						if strings.ToLower(to) != "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
//							break
//						}
//					}
//				}
//				if coin.Name != "usdt-erc20" && coin.Name != "eth" {
//					to = toAddrs[0]
//				}
//
//				//for _, to := range toAddrs {
//				applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
//					Address:     to,
//					AddressFlag: "to",
//					Status:      0,
//					Lastmodify:  cltApply.Lastmodify,
//				})
//				//}
//				for _, from := range fromAddrs {
//					/*
//						func：添加冻结时间限制
//						auth：flynn
//						date： 2020-07-02
//					*/
//					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
//						Address:     from,
//						AddressFlag: "from",
//						Status:      0,
//						Lastmodify:  cltApply.Lastmodify,
//					})
//				}
//				appId, err := cltApply.TransactionAdd(applyAddresses)
//				if err == nil {
//					//开始请求钱包服务归集
//					orderReq := &transfer.bscCollectReq{}
//					orderReq.ApplyId = appId
//					orderReq.OuterOrderNo = cltApply.OutOrderid
//					orderReq.OrderNo = cltApply.OrderId
//					orderReq.MchId = int64(mchId)
//					orderReq.MchName = mchName
//					orderReq.CoinName = c.coinName
//					orderReq.FromAddrs = collectAddrs
//					orderReq.ToAddr = to
//					if coin.Name != c.coinName { //如果是代币归集
//						orderReq.ContractAddr = coin.Token
//						orderReq.Decimal = coin.Decimal
//					}
//					if err := c.walletServerCollect(orderReq); err != nil {
//						log.Errorf("err : %v", err)
//					}
//				}
//			}
//			//}(coin)
//		}
//		//如果需要打手续费
//		if len(pendingFeeTx) > 0 {
//			//生成手续费订单
//			feeApply := &entity.FcTransfersApply{
//				Username:   "Robot",
//				CoinName:   c.coinName,
//				Department: "blockchains-go",
//				OutOrderid: fmt.Sprintf("FEE_%d", time.Now().Nanosecond()),
//				OrderId:    util.GetUUID(),
//				Applicant:  mchName,
//				Operator:   "Robot",
//				AppId:      mchId,
//				Type:       "fee",
//				Purpose:    "自动归集",
//				Status:     int(entity.ApplyStatus_Fee), //因为是即时归集，所以直接把状态置为构建成功
//				Createtime: time.Now().Unix(),
//				Lastmodify: util.GetChinaTimeNow(),
//				Source:     1,
//			}
//			//查找手续费地址
//			//todo 商户手续费告警
//			amt, err := decimal.NewFromString(feeAddress.Amount)
//			if err != nil {
//				//log.Errorf("%v", err)
//				return fmt.Errorf("%v", err)
//			}
//			if amt.LessThan(decimal.NewFromFloat(c.cfg.AlarmFee)) {
//				//log.Errorf("alarm fee %v", amt)
//				if mchName == "hoo" {
//					ErrDingBot.NotifyStr(fmt.Sprintf("商户:%s\n手续费地址:%s\n当前手续费:%s\n手续费报警阈值:%f",
//						mchName, feeAddress.Address, feeAddress.Amount, c.cfg.AlarmFee))
//				}
//				return fmt.Errorf("alarm fee %v", amt)
//			}
//			applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
//			applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
//				Address:     feeAddress.Address,
//				AddressFlag: "from",
//				Status:      0,
//			})
//			feeAddrs := make([]string, 0)
//			for feeAddr := range pendingFeeTx {
//				if c.isUnFreezeAddress(feeAddr) {
//					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
//						Address:     feeAddr,
//						AddressFlag: "to",
//						Status:      0,
//					})
//					feeAddrs = append(feeAddrs, feeAddr)
//				}
//			}
//			appId, err := feeApply.TransactionAdd(applyAddresses)
//			if err == nil {
//				//开始请求钱包服务归集
//				orderReq := &transfer.bscTransferFeeReq{}
//				orderReq.ApplyId = appId
//				orderReq.OuterOrderNo = feeApply.OutOrderid
//				orderReq.OrderNo = feeApply.OrderId
//				orderReq.MchId = int64(mchId)
//				orderReq.MchName = mchName
//				orderReq.CoinName = c.coinName
//				orderReq.FromAddr = feeAddress.Address
//				orderReq.ToAddrs = feeAddrs
//				orderReq.NeedFee = decimal.NewFromFloat(c.cfg.NeedFee).Shift(18).String() //eth -> wei
//				if err := c.walletServerFee(orderReq); err != nil {
//					log.Errorf("err : %v", err)
//				}
//			}
//		}
//		//wg.Wait()
//
//	return nil
//}
//
////创建交易接口参数
//func (c *CollectbscJob) walletServerCollect(orderReq *transfer.bscCollectReq) error {
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/collect", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
//	if err != nil {
//		return err
//	}
//	dd, _ := json.Marshal(orderReq)
//	log.Infof("%s Collect send :%s", c.coinName, string(dd))
//	log.Infof("%s Collect resp :%s", c.coinName, string(data))
//	result := transfer.DecodeWalletServerRespOrder(data)
//	if result == nil {
//		return fmt.Errorf("walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
//	}
//	if result.Code != 0 {
//		log.Error(result)
//		return fmt.Errorf("walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
//	}
//	return nil
//}
//
////创建交易接口参数
//func (c *CollectbscJob) walletServerFee(orderReq *transfer.bscTransferFeeReq) error {
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/fee", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
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
//func (c *CollectbscJob) needTransferFee(addresses []string, minAmount float64) bool {
//	req := transfer.bscBalanceReq{
//		Address: addresses[0],
//	}
//	if len(addresses) > 1 {
//		req.ContractAddr = addresses[1]
//	}
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/balance", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, req)
//	if err != nil {
//		return false
//	}
//	result := transfer.DecodeWalletServerRespOrder(data)
//	if result == nil {
//		return false
//	}
//	if result.Code != 0 || result.Data == nil {
//		log.Error(result)
//		return false
//	}
//	if err != nil {
//		return false
//	}
//	amount, err := decimal.NewFromString(result.Data.(string))
//	if err != nil {
//		return false
//	}
//	minAmt := decimal.NewFromFloat(minAmount)
//	if amount.Shift(-18).Cmp(minAmt) < 0 {
//		return true
//	}
//	return false
//}
//
//func (c *CollectbscJob) isUnFreezeAddress(address string) bool {
//	//v2
//	isContainAddress := false
//	isUnFreeze := false
//	c.limitMap.Range(func(key, value interface{}) bool {
//		lastTxTime := value.(int64)
//		//添加16分钟冻结时间
//		freezeTime := time.Unix(lastTxTime, 0).Add(time.Minute * time.Duration(20)).Unix()
//		now := time.Now().Unix()
//		if now >= freezeTime {
//			if key.(string) == address {
//				isContainAddress = true
//				log.Infof("address=[%s]已解冻,可以进行出账", address)
//				c.limitMap.Store(address, now) //	添加新的冻结时间
//				isUnFreeze = true
//			} else {
//				//从map中移除
//				c.limitMap.Delete(key)
//			}
//		} else {
//			if key.(string) == address {
//				isContainAddress = true
//				TimeStr := func(timestamp int64) string {
//					var timeLayout = "2006-01-02 T 15:04:05.000"
//					return time.Unix(timestamp, 0).Format(timeLayout)
//				}
//				log.Infof("address=[%s]被冻结，解冻时间为：[%s],当前时间为：[%s]", address, TimeStr(freezeTime), TimeStr(now))
//				isUnFreeze = false
//			}
//		}
//		return true
//	})
//
//	if !isContainAddress {
//		c.limitMap.Store(address, time.Now().Unix())
//		isUnFreeze = true
//	}
//	return isUnFreeze
//}
//func (c *CollectbscJob) IsToManyPendingTx(address string) bool {
//	orders, err := entity.FcOrder{}.FindOrders(builder.Eq{"coin_name": c.coinName, "from_address": address, "status": 4}.
//		And(builder.Like{"outer_order_no", "FEE_"}), 50)
//	if err != nil {
//		log.Errorf("查询fc_order error,Err=%v", err)
//		return false
//	}
//	var txids []string
//	for _, order := range orders {
//		if order.TxId != "" {
//			txids = append(txids, order.TxId)
//		}
//	}
//	pushes, err := entity.FcTransPush{}.FindTransPush(builder.Eq{"from_address": address}.And(builder.In("transaction_id", txids)))
//	if err != nil {
//		log.Errorf("查询fc_trans_push error,Err=%v", err)
//		return false
//	}
//	pendingNum := len(txids) - len(pushes)
//	if pendingNum >= 30 {
//		log.Infof("totalNum:[%d],pendingNum:=[%d]", len(txids), pendingNum)
//		return true
//	}
//	log.Infof("total find number=[%d],pending number=[%d]", len(txids), pendingNum)
//	return false
//}
//
//type BalanceStruct struct {
//	Status  string          `json:"status"`
//	Message string          `json:"message"`
//	Result  decimal.Decimal `json:"result"`
//	//"status": "1",
//	//"message": "OK",
//	//"result": "4009811415661147191"
//}
//
//func getTokenBalance(contractaddress, addr string) (decimal.Decimal, error) {
//	apikey := "MXKM5DKHND1KUGKF3PPIDQQJXC2IRIDVUV"
//	url := "https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest&apikey=%s"
//	if strings.TrimSpace(addr) == "" {
//		return decimal.Zero, errors.New("empty addr blanance")
//	}
//	if contractaddress != "" {
//		url = fmt.Sprintf(url, addr, apikey)
//	} else {
//		url = "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
//		url = fmt.Sprintf(url, contractaddress, addr, apikey)
//	}
//
//	resultData, err := util.Get(url)
//	if err != nil {
//		log.Error(string(resultData))
//		return decimal.Zero, err
//	}
//	result := new(BalanceStruct)
//	err = json.Unmarshal(resultData, result)
//	if err != nil {
//		log.Error(string(resultData))
//		return decimal.Zero, err
//	}
//
//	if result.Status != "1" {
//		return decimal.Zero, errors.New(string(resultData))
//	}
//	return result.Result, nil
//}
