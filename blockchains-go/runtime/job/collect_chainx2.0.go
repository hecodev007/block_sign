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
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type CollectChainXJob struct {
	coinName string
	cfg      conf.Collect2
}

func NewCollectChainXJob(cfg conf.Collect2) cron.Job {
	return CollectChainXJob{
		coinName: "pcx",
		cfg:      cfg,
	}
}

func (c CollectChainXJob) Run() {
	var (
		mchs []*entity.FcMch
		err  error
	)
	stChainXt := time.Now()

	log.Infof("*** %s collect task start***", c.coinName)
	defer log.Infof("*** %s collect task end, use time : %f s ", c.coinName, time.Since(stChainXt).Seconds())
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
func (c *CollectChainXJob) collect(mchId int, mchName string) error {
	//获取币种的配置
	ChainXCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(ChainXCoins) == 0 {
		return errors.New("do not find fis coin")
	}
	//只有一个coin
	coin := ChainXCoins[0]
	//获取有余额的地址
	fromAddrs, err1 := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchId}.
		And(builder.Expr("amount > ? and forzen_amount = 0", c.cfg.MinAmount)), c.cfg.MaxCount)
	if err1 != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err1.Error())
	}
	if len(fromAddrs) == 0 {
		return nil
		//return fmt.Errorf("%s don't have need collected from address", mchName)
	}
	//获取归集的目标冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   c.coinName,
	})
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s don't have cold address", mchName)
	}
	fee := decimal.NewFromFloat(c.cfg.NeedFee)
	var (
		totalNum, successNum, failNum int
	)
	totalNum = len(fromAddrs)
	for _, from := range fromAddrs {
		amount, _ := decimal.NewFromString(from.Amount)
		//减去手续费
		amount = amount.Sub(fee)
		if amount.LessThanOrEqual(decimal.NewFromInt(0)) {
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
			Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
			Createtime: time.Now().Unix(),
		}
		if coin.Name != c.coinName {
			cltApply.Eostoken = coin.Token
			cltApply.Eoskey = coin.Name
		}
		to := toAddrs[0] //随机取个地址
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
			orderReq := &transfer.ChainXOrderRequest{}
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(mchId)
			orderReq.MchName = mchName
			orderReq.CoinName = c.coinName

			orderReq.FromAddress = from.Address
			orderReq.ToAddress = to
			orderReq.Amount = amount.Shift(int32(coin.Decimal)).String()
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
				Quantity:     orderReq.Amount,
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
			orderHot.TxId = txid
			orderHot.Status = int(status.BroadcastStatus)
			//保存热表
			err = dao.FcOrderHotInsert(orderHot)
			if err != nil {
				err = fmt.Errorf("[%s]归集保存订单[%s]数据异常:[%s]", c.coinName, orderHot.OuterOrderNo, err.Error())
				//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
				log.Error(err.Error())
			}
		} else {
			failNum++
		}
	}
	log.Infof("%s 归集完成，总共需要归集笔数： %d，成功笔数： %d，失败笔数：%d", strings.ToUpper(c.coinName), totalNum, successNum, failNum)
	return nil
}

func (c *CollectChainXJob) walletServerCreateHot(orderReq *transfer.ChainXOrderRequest) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", c.cfg.Url, c.coinName), c.cfg.User, c.cfg.Password, orderReq)
	if err != nil {
		return "", fmt.Errorf("%s collect fail,,from=[%s],to=[%s],amount=[%s],err=%v", orderReq.CoinName, orderReq.FromAddress,
			orderReq.ToAddress, orderReq.Amount, err)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", orderReq.CoinName, string(dd))
	log.Infof("%s Collect resp :%s", orderReq.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.(string), nil
}