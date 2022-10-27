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

//主要是归集qtum的代币

type CollectQtumJob struct {
	coinName string
	cfg      conf.Collect2
	limitMap sync.Map
}

func NewCollectQtumJob(cfg conf.Collect2) cron.Job {
	return CollectQtumJob{
		coinName: "qtum",
		cfg:      cfg,
		limitMap: sync.Map{}, //初始化限制表
	}
}

func (c CollectQtumJob) Run() {
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

func (c *CollectQtumJob) collect(mchId int, mchName string) error {
	//先获取bnb地址
	qutmCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(qutmCoins) == 0 {
		return errors.New("do not find qtum coin")
	}
	//获取所有合约地址
	pid := qutmCoins[0].Id
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err != nil {
		return err
	}
	//手续费交易
	pendingFeeTx := make(map[string]bool)
	if len(coins) > 0 {
		//查找手续费地址
		changeAddress := &entity.FcAddressAmount{}
		has, err := changeAddress.Get(builder.In("address",
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
		if err != nil || !has {
			log.Errorf("查找找零地址错误:%v", err)
			return fmt.Errorf("查找找零地址错误:%v", err)
		}

		//只处理代币归集，不处理qtum的归集
		for _, coin := range coins {
			log.Infof("商户：%s,执行币种：%s", mchName, coin.Name)
			toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
				"type":        address.AddressTypeCold,
				"status":      address.AddressStatusAlloc,
				"platform_id": mchId,
				"coin_name":   c.coinName,
			})
			if err != nil || len(toAddrs) == 0 {
				continue
			}
			thresh := c.cfg.MinAmount
			//获取有余额的地址
			fAA, err1 := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >=  ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)
			if err1 != nil {
				log.Errorf("查询归集数据异常:%v", err1)
				continue
			}
			if len(fAA) == 0 {
				continue
			}
			collectAddrs := make([]*entity.FcAddressAmount, 0)

			for _, fromAddr := range fAA {
				//判断代币转账是否需要转手续费
				if c.needTransferFee(fromAddr.Address, c.cfg.NeedFee) {
					pendingFeeTx[fromAddr.Address] = true
				} else {
					//直接归集
					collectAddrs = append(collectAddrs, fromAddr)
				}
			}
			// 先对代币进行归集
			if len(collectAddrs) > 0 {
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
						orderReq := &transfer.QtumOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = cltApply.OutOrderid
						orderReq.OrderNo = cltApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName

						oa := transfer.QtumOrderAddressReq{
							Address:  to,
							Amount:   amount.Shift(int32(coin.Decimal)).IntPart(),
							Quantity: amount.Shift(int32(coin.Decimal)).String(),
						}
						orderReq.FromAddress = from.Address
						orderReq.ChangeAddress = changeAddress.Address
						orderReq.Token = coin.Token
						orderReq.OrderAddress = append(orderReq.OrderAddress, oa)

						if err := c.walletServerCreateCold(orderReq); err != nil {
							if strings.Contains(err.Error(), "needFee") {
								pendingFeeTx[from.Address] = true
							}
							log.Errorf("err : %v", err)
						} else {
							log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s]",
								strings.ToUpper(coin.Name), mchId, appId, from.Address, to, amount.String())
							time.Sleep(10 * time.Second)
						}
					}

				}
			}
		}
		//打手续费
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
					ErrDingBot.NotifyStr(fmt.Sprintf("%s 手续费不足报警,当前手续费=%v", strings.ToUpper(c.coinName), amt))
					log.Errorf("%s 手续费不足报警,当前手续费=%v", strings.ToUpper(c.coinName), amt)
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
						orderReq := &transfer.QtumOrderRequest{}
						orderReq.ApplyId = appId
						orderReq.OuterOrderNo = feeApply.OutOrderid
						orderReq.OrderNo = feeApply.OrderId
						orderReq.MchId = int64(mchId)
						orderReq.MchName = mchName
						orderReq.CoinName = c.coinName

						oa := transfer.QtumOrderAddressReq{
							Address:  feeAddr,
							Amount:   decimal.NewFromFloat(c.cfg.NeedFee).Shift(8).IntPart(),
							Quantity: decimal.NewFromFloat(c.cfg.NeedFee).Shift(8).String(),
						}
						orderReq.FromAddress = feeAddress.Address
						orderReq.ChangeAddress = feeAddress.Address
						orderReq.OrderAddress = append(orderReq.OrderAddress, oa)
						if err := c.walletServerCreateCold(orderReq); err != nil {
							log.Errorf("err : %v", err)
						} else {
							log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%f]",
								strings.ToUpper(c.coinName), mchId, appId, feeAddress.Address, feeAddr, c.cfg.NeedFee)
							//休眠10秒
							time.Sleep(10 * time.Second)
						}
					}
				}
			}
		}
	}
	return nil
}

func (c *CollectQtumJob) needTransferFee(address string, minAmount float64) bool {
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(
		builder.Expr("type =? and address=? and coin_type=? and amount<?", 2, address, c.coinName, minAmount), 1)
	if err != nil {
		log.Errorf("get fee address balance error,err=%v", err)
		return false
	}
	if len(fromAddrs) > 0 {
		//需要手续费
		return true
	}
	return false
}

func (c *CollectQtumJob) walletServerCreateCold(orderReq *transfer.QtumOrderRequest) error {
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%d],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.OrderAddress[0].Address, orderReq.OrderAddress[0].Amount, err)
	}
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("%s walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.CoinName, orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		if result.Code == 40001 && strings.Contains(result.Message, "vaild utxo no enough") {
			return errors.New("needFee")
		} else {
			log.Error(result)
			return fmt.Errorf("%s walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.CoinName, orderReq.OuterOrderNo, string(data))
		}

	}
	return nil
}
