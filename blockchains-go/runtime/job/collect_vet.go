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
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectVetJob struct {
	coinName string
	feeName  string
	cfg      conf.Collect2
}

const (
	FeeDecimal         int32 = 18
	FeeContractAddress       = "0x0000000000000000000000000000456E65726779"
)

func NewCollectVetJob(cfg conf.Collect2) cron.Job {
	return CollectVetJob{
		coinName: "vet",
		feeName:  "vtho",
		cfg:      cfg,
	}
}

func (c CollectVetJob) Run() {
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

func (c *CollectVetJob) collect(mchId int, mchName string) error {
	//先获取vet地址
	VetCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(VetCoins) == 0 {
		return errors.New("do not find Vet coin")
	}
	//获取所有合约地址
	pid := VetCoins[0].Id
	AllCoins, err1 := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}).And(builder.NotIn("name", c.cfg.IgnoreCoins)))
	if err1 != nil {
		return err1
	}
	var (
		coins []*entity.FcCoinSet
	)
	coins = append(coins, VetCoins...)
	if len(AllCoins) > 0 {
		for _, coin := range AllCoins {
			coins = append(coins, coin)
		}
	}
	log.Infof("执行币种数量:%d", len(coins))
	if len(coins) > 0 {
		feeAddress, err := entity.FcAddressAmount{}.FindAddressAndAmount(
			builder.Expr("type =? and coin_type=? and amount>=?", 1, c.feeName, c.cfg.NeedFee), 5)
		if err != nil || len(feeAddress) == 0 {
			return fmt.Errorf("do not find any fee address,err=[%v]", err)
		}
		feeA := feeAddress[rand.Intn(len(feeAddress))]
		feeAmount, _ := decimal.NewFromString(feeA.Amount)
		numsDec := feeAmount.Div(decimal.NewFromInt(100))
		if feeAmount.LessThan(decimal.NewFromFloat(c.cfg.AlarmFee)) {
			ErrDingBot.NotifyStr(fmt.Sprintf("VTHO 手续费不足报警,当前手续费=%d mchID=%d", feeAmount.IntPart(), mchId))
		}
		nums := int(numsDec.IntPart())
		log.Infof("总共可执行归集数量为： %d", nums)
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
			thresh := 1.0
			if coin.Name == c.coinName {
				thresh = c.cfg.MinAmount
			} else if coin.Name == c.feeName {
				thresh = c.cfg.NeedFee
			}

			//获取有余额的地址
			fAA, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": coin.Name, "app_id": mchId}.
				And(builder.Expr("amount >  ? and forzen_amount = 0", thresh)).
				And(builder.NotIn("address", toAddrs)), c.cfg.MaxCount)

			if err != nil {
				log.Errorf("查询归集数据异常:%s", err.Error())
				continue
			}
			if len(fAA) == 0 {
				//log.Errorf("%s don't have need collected address", mchName)
				continue
			}
			for _, from := range fAA {
				if nums <= 0 {
					return errors.New("手续费不足以支撑币种归集，退出本轮归集")
				}
				//需要判断是否有足够的手续费
				//判断代币转账是否需要转手续费
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
					fee := decimal.NewFromFloat(c.cfg.NeedFee).Shift(int32(coin.Decimal))
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
					//填充参数
					orderReq := &transfer.VetRepay{}

					orderReq.From = from.Address
					orderReq.To = to
					orderReq.Repay = feeA.Address
					orderReq.Amount = amount.String()
					createData, _ := json.Marshal(orderReq)

					orderHot := &entity.FcOrderHot{
						ApplyId:      int(appId),
						ApplyCoinId:  coin.Id,
						OuterOrderNo: cltApply.OutOrderid,
						OrderNo:      cltApply.OrderId,
						MchName:      mchName,
						CoinName:     c.coinName,
						FromAddress:  from.Address,
						ToAddress:    to,
						Amount:       amount.IntPart(), //转换整型
						Quantity:     amount.String(),
						Decimal:      int64(coin.Decimal),
						CreateData:   string(createData),
						Status:       int(status.UnknowErrorStatus),
						CreateAt:     time.Now().Unix(),
						UpdateAt:     time.Now().Unix(),
					}

					txid, errTX := c.walletServerRepayTransferHot(orderReq, coin.Name)
					if errTX != nil {
						orderHot.Status = int(status.BroadcastErrorStatus)
						orderHot.ErrorMsg = errTX.Error()
						dao.FcOrderHotInsert(orderHot)
						log.Errorf("归集错误，err : %v", errTX)
						continue
					} else {
						nums--
						orderHot.TxId = txid
						orderHot.Status = int(status.BroadcastStatus)
						log.Infof("成功归集一笔%s,MchId=[%d],ApplyId=[%d],from=[%s],to=[%s],amount=[%s],txid=[%s]",
							strings.ToUpper(coin.Name), mchId, appId, from.Address, to, amount.String(), txid)
						//保存热表
						err = dao.FcOrderHotInsert(orderHot)
						if err != nil {
							err = fmt.Errorf("[%s]归集保存订单[%s]数据异常:[%s]", c.coinName, orderHot.OuterOrderNo, err.Error())
							//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
							log.Error(err.Error())
						}
					}
				}
			}
		}
	}
	return nil
}

func (c *CollectVetJob) walletServerRepayTransferHot(orderReq *transfer.VetRepay, coinName string) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/repay", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("post data to service error,Err=[%v]", err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", coinName, string(data))
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.New("json unmarshal response data error")
	}
	if resp["error"] != nil {
		return "", fmt.Errorf("vet transfer error,Err=%v", resp["error"])
	}
	return resp["result"].(string), nil
}
