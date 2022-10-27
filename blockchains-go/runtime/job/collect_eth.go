package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/onrik/ethrpc"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

//黑名单zild地址 不归集
var errZildArr = []string{
	"0x00f76cac3deadc6c077290f01ce41e121c527c9f",
	"0x004b41d01048439adc6c5c6e7f19866d0350cdd3",
	"0x002ee42b5328f2b7d34e0b4928d09ea094919a9b",
	"0x002ee42b5328f2b7d34e0b4928d09ea094919a9b",
	"0x000fed2781a5cd93e4aa4c3d56c611e19742cf32",
	"0x00721e3bae57c259547b13a42b401193ba69992c",
	"0x002da1952d0959f4d4440d3797b65b68a0001d55",
	"0x009a3c96e0d3f6258d29eea94a485896b65fdeb0",
	"0x001d521d2d7ae10d122b4a2af3089ac4ba279f8a",
	"0x006fc567f5a12ea5fa1f7c5f36874d56155ca625",
	"0x00cd2c88343f1422a42d04e6efbbbc2ff5ef2ca2",
	"0x00e7437d7fe5cb4cf626677c1a47a1514835803d",
}

type CollectEthJob struct {
	coinName                 string
	cfg                      conf.Collect2
	limitMap                 sync.Map
	feeAddrPendingOutOrderId []string
	ethRpc                   *ethrpc.EthRPC
}

func NewCollectEthJob(cfg conf.Collect2) cron.Job {
	//client := ethrpc.New("http://127.0.0.1:8545")
	client := ethrpc.New(cfg.Node)
	version, err := client.Web3ClientVersion()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("eth version[%s]", version)
	return CollectEthJob{
		coinName: "eth",
		cfg:      cfg,
		ethRpc:   client,
		limitMap: sync.Map{}, //初始化限制表
	}
}

func getFromFixAddr(models []*entity.FcFixAddress, addr string) *entity.FcFixAddress {
	for _, model := range models {
		if strings.ToLower(addr) == strings.ToLower(model.Address) {
			return model
		}
	}
	return nil
}

func (c CollectEthJob) Run() {
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

func (c *CollectEthJob) collect(mchId int, mchName string) error {
	//添加获取pending交易的方法
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
	if err != nil || !has {
		return fmt.Errorf("find fee address error ,Err=%v", err)
	}

	//todo 这个方法好像有点问题
	//if c.IsToManyPendingTx(feeAddress.Address) {
	//	return fmt.Errorf(" %s have to many pending tx", feeAddress.Address)
	//}
	//最新nonce

	//start := time.Now()
	//log.Infof("=== %s collect task start ===", mchName)
	//defer log.Infof("=== %s collect task end, use time : %f s ===", mchName, time.Since(start).Seconds())
	//获取所有eth币种信息
	var (
		coins []*entity.FcCoinSet
	)
	if len(c.cfg.AssignCoins) > 0 {
		coins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": 5}.Or(builder.Eq{"id": 5})).And(builder.In("name", c.cfg.AssignCoins)))
	} else {
		coins, err = entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": 5}.Or(builder.Eq{"id": 5})).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	}
	if err != nil {
		return err
	}

	fixAddrList, err := dao.FcFixAddressActiveList()
	if err != nil {
		log.Errorf("[固定地址金额]执行dao.FcFixAddressActiveList出错 %v", err)
	}

	if len(coins) > 0 {
		//wg := &sync.WaitGroup{}
		pendingFeeTx := make(map[string]bool)
		for _, coin := range coins {
			if "mof" == strings.ToLower(coin.Name) {
				log.Infof("代币[MOF]使用自己的独立归集程序，跳过")
				continue
			}

			//查看是否需要归集
			if coin.IsCollect == 0 {
				continue
			}
			log.Infof("代币：%s,归集开启", coin.Name)

			//go func(coin *entity.FcCoinSet) {
			//wg.Add(1)
			//defer wg.Done()
			//log.Infof("--- %s 归集任务 start ---",coin.Name)
			//defer 	log.Infof("--- %s 归集任务 end ---",coin.Name)
			//获取归集的目标冷地址
			toAddrs := make([]string, 0)
			if mchId == 1 {
				toAddrs = append(toAddrs, "0x0093e5f2a850268c0ca3093c7ea53731296487eb")
				toAddrs = append(toAddrs, "0x002471c86e9e97d393d84bddfa7d555a7fa2917a")
				toAddrs = append(toAddrs, "0x0055e75217ca5cb5aa8290cd966f9d85751a7993")
			} else {
				toAddrs, err = entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
					"type":        address.AddressTypeCold,
					"status":      address.AddressStatusAlloc,
					"platform_id": mchId,
					"coin_name":   c.coinName,
				}.And(builder.Neq{"address": "0x9760862d09b70433a91ff27cbd069f51ef1cbd5c"}))
				if err != nil || len(toAddrs) == 0 {
					//fmt.Errorf("%s don't hava cold address", mchName)
					continue
				}
			}

			thresh := 1000.0
			if strings.ToLower(coin.Name) == c.coinName {
				thresh = c.cfg.MinAmount
			} else if strings.ToLower(coin.Name) == "yfi" {
				thresh = 0.01
			} else if strings.ToLower(coin.Name) == "dxd" {
				thresh = 0.01
			}
			//数据库设置金额
			collectThreshold, _ := decimal.NewFromString(coin.CollectThreshold)
			collectThresholdFloat, _ := collectThreshold.Float64()
			if collectThresholdFloat <= 0 {
				log.Infof("代币：%s,没有设置参数，使用默认金额：%v", coin.Name, thresh)
			} else {
				thresh = collectThresholdFloat
			}

			var fromAddrs []string
			//获取有余额的地址
			if strings.ToLower(coin.Name) == c.coinName {
				// 不要归集固定金额的地址的ETH余额
				notInAddrs := toAddrs
				for _, fixAddr := range fixAddrList {
					notInAddrs = append(notInAddrs, fixAddr.Address)
				}
				//fixAddrs, err := dao.FcFindFixAddressList()
				//if err != nil {
				//	log.Errorf("[固定地址金额] dao.FcFindFixAddressList 执行失败 %v", err)
				//}
				log.Infof("[固定地址金额] NOT IN address %v", notInAddrs)
				// 归集主链币
				fromAddrs, err = entity.FcAddressAmount{}.FindAddress(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
					And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
					And(builder.NotIn("address", notInAddrs)), c.cfg.MaxCount)
			} else {
				fromAddrs, err = entity.FcAddressAmount{}.FindAddress(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
					And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)).
					And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			}

			if err != nil {
				log.Errorf("查询归集数据异常:%s", err.Error())
				continue
			}
			if len(fromAddrs) == 0 {
				log.Errorf("%s don't hava need collected address", mchName)
				continue
			}
			collectAddrs := make([]string, 0)
			//feeAddrs := make([]string, 0)
			if strings.ToLower(coin.Name) != c.coinName {
				//如果是代币归集，那么我们还需要考虑是否足够的eth手续费
				//过滤出来需要打手续费的地址
				for _, fromAddr := range fromAddrs {
					if strings.ToLower(coin.Name) == "zild" {
						if util.IsInArrayStr(strings.ToLower(fromAddr), errZildArr) {
							//异常地址不归集
							log.Errorf("币种：%s,黑名单地址：%s,略过", coin.Name, fromAddr)
							continue
						}
					}
					if c.needTransferFee([]string{fromAddr}, c.cfg.NeedFee) {
						if _, ok := pendingFeeTx[fromAddr]; !ok {
							pendingFeeTx[fromAddr] = true
						}
					} else {
						time.Sleep(500 * time.Millisecond)
						//检查代币了链上金额是否大于0
						tokenBalance, err := getTokenBalance(coin.Token, fromAddr)
						if err != nil {
							log.Errorf("检查代币金额失败：%s", err.Error())
							continue
						}
						if tokenBalance.IsZero() {
							continue
						}
						log.Infof("地址：%s,token:%s,精度：%d,查询token【%s】,实际金额：%s", fromAddr, coin.Name, coin.Decimal, coin.Token, tokenBalance.Shift(int32(-1*coin.Decimal)).String())
						collectAddrs = append(collectAddrs, fromAddr)

					}
				}
			} else {
				//eth 归集
				collectAddrs = fromAddrs
			}
			if len(collectAddrs) > 0 {
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

				//减少某个地址的归集次数，因为被浏览器标记了
				if strings.ToLower(to) == "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
					//再次随机两次
					for r := 0; r < 2; r++ {
						to = toAddrs[rand.Intn(randNum)]
						if strings.ToLower(to) != "0x980a4732c8855ffc8112e6746bd62095b4c2228f" {
							break
						}
					}
				}
				if coin.Name != "usdt-erc20" && coin.Name != "eth" {
					to = toAddrs[0]
				}

				//for _, to := range toAddrs {
				applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
					Address:     to,
					AddressFlag: "to",
					Status:      0,
					Lastmodify:  cltApply.Lastmodify,
				})
				//}
				for _, from := range fromAddrs {
					/*
						func：添加冻结时间限制
						auth：flynn
						date： 2020-07-02
					*/
					applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
						Address:     from,
						AddressFlag: "from",
						Status:      0,
						Lastmodify:  cltApply.Lastmodify,
					})
				}
				appId, err := cltApply.TransactionAdd(applyAddresses)
				if err == nil {
					//开始请求钱包服务归集
					orderReq := &transfer.EthCollectReq{}
					orderReq.ApplyId = appId
					orderReq.OuterOrderNo = cltApply.OutOrderid
					orderReq.OrderNo = cltApply.OrderId
					orderReq.MchId = int64(mchId)
					orderReq.MchName = mchName
					orderReq.CoinName = c.coinName
					orderReq.FromAddrs = collectAddrs
					orderReq.ToAddr = to
					if coin.Name != c.coinName { //如果是代币归集
						orderReq.ContractAddr = coin.Token
						orderReq.Decimal = coin.Decimal
					}
					if err := c.walletServerCollect(orderReq); err != nil {
						log.Errorf("err : %v", err)
					}
				}
			}
			//}(coin)
		}
		//如果需要打手续费
		if len(pendingFeeTx) > 0 {
			//if c.IsToManyPendingTx(feeAddress.Address) {
			//	return fmt.Errorf(" %s have to many pending tx", feeAddress.Address)
			//}
			coininfo, err := dao.FcCoinSetGetByName(c.coinName, 1)
			if err != nil {
				log.Errorf("获取eth币种信息错误：%s", err.Error())
				return err
			}

			needLatest := false

			if !c.cfg.UseLatestNonce {
				feeLatestNonce, err := c.ethRpc.EthGetTransactionCount(feeAddress.Address, "latest")
				if err != nil {
					log.Errorf("获取手续费地址【%s】的latest nonce错误", feeAddress.Address)
					return err
				}

				//pending nonce
				feePendingNonce, err := c.ethRpc.EthGetTransactionCount(feeAddress.Address, "pending")
				if err != nil {
					log.Errorf("获取手续费地址【%s】的pending nonce错误", feeAddress.Address)
					return err
				}
				if feePendingNonce-feeLatestNonce > 10 {
					log.Infof("手续费地址【%s】链上pending笔数过多,启动Latest重置，数量：%d", feeAddress.Address, feePendingNonce-feeLatestNonce)
					needLatest = true
				}
				log.Infof("目前手续费地址【%s】,pending笔数约为：%d", feeAddress.Address, feePendingNonce-feeLatestNonce)
			}
			//查找手续费地址
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

			log.Infof("预期手续费笔数：%d", len(pendingFeeTx))
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
				feeAddrs := make([]string, 0)
				feeApply.OutOrderid = fmt.Sprintf("FEE_%d", time.Now().Nanosecond())
				feeApply.OrderId = util.GetUUID()
				applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
					Address:     feeAddr,
					AddressFlag: "to",
					Status:      0,
				})
				feeAddrs = append(feeAddrs, feeAddr)
				appId, err := feeApply.TransactionAdd(applyAddresses)

				feeAmt := decimal.NewFromFloat(c.cfg.NeedFee).Shift(18).String()
				fixAddr := getFromFixAddr(fixAddrList, feeAddr)
				if fixAddr != nil {
					if fixAddr.Payed == 0 {
						log.Infof("[固定地址金额] 地址 %s 需要打手续费 %s", feeAddr, fixAddr.Amount)
						fixAddrAmount, _ := decimal.NewFromString(fixAddr.Amount)
						feeAmt = fixAddrAmount.Shift(18).String()
						dao.UpdateFcFixAddressById(fixAddr.Id)
					} else {
						log.Infof("[固定地址金额] 地址 %s 需要打手续费，但是此记录之前已处理 本次不特殊处理", feeAddr)
					}
				}

				if err == nil {
					//开始请求钱包服务归集
					orderReq := &transfer.ETHOrderRequest{
						OrderRequestHead: transfer.OrderRequestHead{
							ApplyId:        appId,
							ApplyCoinId:    int64(coininfo.Id),
							OuterOrderNo:   feeApply.OutOrderid,
							OrderNo:        feeApply.OrderId,
							MchName:        mchName,
							CoinName:       c.coinName,
							Worker:         "0",
							RecycleAddress: "",
							Sign:           "",
							CurrentTime:    "",
						},
						FromAddress:     feeAddress.Address,
						ToAddress:       feeAddr,
						Amount:          feeAmt, //目前非动态
						ContractAddress: "",
						Token:           "",
					}
					if needLatest {
						orderReq.Latest = true
					}
					createData, _ := json.Marshal(orderReq)
					order := &entity.FcOrder{
						ApplyId:      int(appId),
						ApplyCoinId:  coininfo.Id,
						OuterOrderNo: orderReq.OuterOrderNo,
						OrderNo:      orderReq.OrderNo,
						MchName:      mchName,
						CoinName:     c.coinName,
						FromAddress:  orderReq.FromAddress,
						ToAddress:    orderReq.ToAddress,
						Amount:       feeAmt, //转换整型
						Quantity:     feeAmt,
						Decimal:      int64(coininfo.Decimal),
						CreateData:   string(createData),
						Status:       int(status.UnknowErrorStatus),
						CreateAt:     time.Now().Unix(),
						UpdateAt:     time.Now().Unix(),
					}
					txid, err := c.walletServerCreateHot(orderReq)
					if err != nil {
						log.Errorf("发送手续费失败 : %s", err.Error())
						order.Status = int(status.BroadcastErrorStatus)
						order.ErrorMsg = err.Error()
						_ = dao.FcOrderInsert(order)
					} else {
						log.Errorf("发送手续费成功，接收地址:[%s] ，txid:[%s]", feeAddr, txid)
						order.TxId = txid
						order.Status = int(status.BroadcastStatus)
						_ = dao.FcOrderInsert(order)
					}
				} else {
					log.Errorf("保存地址【%s】订单失败,error:%s", feeAddr, err.Error())
				}
				time.Sleep(2 * time.Second)
			}
		}
		//wg.Wait()
	}
	return nil
}

//创建交易接口参数
func (c *CollectEthJob) walletServerCollect(orderReq *transfer.EthCollectReq) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/collect", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", c.coinName, string(dd))
	log.Infof("%s Collect resp :%s", c.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//创建交易接口参数
//func (c *CollectEthJob) walletServerFee(orderReq *transfer.EthTransferFeeReq) error {
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

func (c *CollectEthJob) needTransferFee(addresses []string, minAmount float64) bool {
	req := transfer.EthBalanceReq{
		Address: addresses[0],
	}
	if len(addresses) > 1 {
		req.ContractAddr = addresses[1]
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/balance", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, req)
	if err != nil {
		return false
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return false
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return false
	}
	if err != nil {
		return false
	}
	amount, err := decimal.NewFromString(result.Data.(string))
	if err != nil {
		return false
	}
	minAmt := decimal.NewFromFloat(minAmount)
	if amount.Shift(-18).Cmp(minAmt) < 0 {
		return true
	}
	return false
}

func (c *CollectEthJob) isUnFreezeAddress(address string) bool {
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
func (c *CollectEthJob) IsToManyPendingTx(address string) bool {
	orders, err := entity.FcOrder{}.FindOrders(builder.Eq{"coin_name": c.coinName, "from_address": address, "status": 4}.
		And(builder.Like{"outer_order_no", "FEE_"}), 50)
	if err != nil {
		log.Errorf("查询fc_order error,Err=%v", err)
		return false
	}
	var txids []string
	for _, order := range orders {
		if order.TxId != "" {
			txids = append(txids, order.TxId)
		}
	}
	pushes, err := entity.FcTransPush{}.FindTransPush(builder.Eq{"from_address": address}.And(builder.In("transaction_id", txids)))
	if err != nil {
		log.Errorf("查询fc_trans_push error,Err=%v", err)
		return false
	}
	pendingNum := len(txids) - len(pushes)
	if pendingNum >= 10 {
		log.Infof("totalNum:[%d],pendingNum:=[%d]", len(txids), pendingNum)
		return true
	}
	log.Infof("total find number=[%d],pending number=[%d]", len(txids), pendingNum)
	return false
}

type BalanceStruct struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  decimal.Decimal `json:"result"`
	//"status": "1",
	//"message": "OK",
	//"result": "4009811415661147191"
}

func getTokenBalance(contractaddress, addr string) (decimal.Decimal, error) {
	apikey := "MXKM5DKHND1KUGKF3PPIDQQJXC2IRIDVUV"
	url := "https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest&apikey=%s"
	if strings.TrimSpace(addr) == "" {
		return decimal.Zero, errors.New("empty addr blanance")
	}
	if contractaddress == "" {
		url = fmt.Sprintf(url, addr, apikey)
	} else {
		url = "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
		url = fmt.Sprintf(url, contractaddress, addr, apikey)
	}

	resultData, err := util.Get(url)
	if err != nil {
		log.Error(string(resultData))
		return decimal.Zero, err
	}
	result := new(BalanceStruct)
	err = json.Unmarshal(resultData, result)
	if err != nil {
		log.Error(string(resultData))
		return decimal.Zero, err
	}

	if result.Status != "1" {
		return decimal.Zero, errors.New(string(resultData))
	}
	return result.Result, nil
}

func (c *CollectEthJob) PendingTxNum(address string) (int, error) {
	orders, err := entity.FcOrder{}.FindOrders(builder.Eq{"coin_name": c.coinName, "from_address": address, "status": 4}.
		And(builder.Like{"outer_order_no", "FEE_"}), 50)
	if err != nil {
		log.Errorf("查询fc_order error,Err=%v", err)
		return 0, err
	}
	var txids []string
	for _, order := range orders {
		if order.TxId != "" {
			txids = append(txids, order.TxId)
		}
	}
	pushes, err := entity.FcTransPush{}.FindTransPush(builder.Eq{"from_address": address}.And(builder.In("transaction_id", txids)))
	if err != nil {
		log.Errorf("查询fc_trans_push error,Err=%v", err)
		return 0, err
	}
	pendingNum := len(txids) - len(pushes)
	if pendingNum >= 30 {
		log.Infof("totalNum:[%d],pendingNum:=[%d]", len(txids), pendingNum)
		return pendingNum, nil
	}
	log.Infof("total find number=[%d],pending number=[%d]", len(txids), pendingNum)
	return 0, fmt.Errorf(",目前地址【%s】pendingNum数量较多", address)
}

func (c *CollectEthJob) walletServerCreateHot(orderReq *transfer.ETHOrderRequest) (string, error) {
	if c.cfg.UseLatestNonce {
		orderReq.Latest = true
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", "http://192.170.1.46:12011", c.coinName), "hsc", "7c0dc78d742b0bd805635d00a63f39fcd86945fb13575440", orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeHscTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result["data"].(string), nil
}
