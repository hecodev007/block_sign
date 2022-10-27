package recycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"time"
	"xorm.io/builder"
)

type DotRecycleService struct {
	coinName string
}

func NewDotRecycleService() *DotRecycleService {
	return &DotRecycleService{coinName: "dot"}
}

func (c *DotRecycleService) RecycleCoin(mchInfo *entity.FcMch, to string, num int) error {
	//获取币种的配置
	DotCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(DotCoins) == 0 {
		return errors.New("do not find dot coin")
	}
	//只有一个coin
	coin := DotCoins[0]
	minAmount := 1.02
	//获取有余额的地址
	fromAddrs, err1 := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 2, "coin_type": c.coinName, "app_id": mchInfo.Id}.
		And(builder.Expr("amount > ? and forzen_amount = 0", minAmount)), num)
	if err1 != nil {
		//log.Errorf("查询归集数据异常:%s", err.Error())
		return fmt.Errorf("查询归集数据异常:%s", err1.Error())
	}
	if len(fromAddrs) == 0 {
		return nil
	}
	fee := decimal.NewFromFloat(0.02)
	for _, from := range fromAddrs {
		amount, _ := decimal.NewFromString(from.Amount)
		//减去手续费
		amount = amount.Sub(fee)
		if amount.LessThanOrEqual(decimal.NewFromInt(0)) {
			continue
		}
		//生产归集订单
		cltApply := &entity.FcTransfersApply{
			Username:   "Robot",
			Department: "blockchains-go",
			Applicant:  mchInfo.Platform,
			OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
			OrderId:    util.GetUUID(),
			Operator:   "Robot",
			CoinName:   c.coinName,
			Type:       "gj",
			Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
			Lastmodify: util.GetChinaTimeNow(),
			AppId:      mchInfo.Id,
			Source:     1,
			Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
			Createtime: time.Now().Unix(),
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
			orderReq := &transfer.DotOrderRequest{}
			orderReq.ApplyId = appId
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(mchInfo.Id)
			orderReq.MchName = mchInfo.Platform
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
				MchName:      mchInfo.Platform,
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
				orderHot.Status = int(status.BroadcastErrorStatus)
				orderHot.ErrorMsg = err.Error()
				dao.FcOrderHotInsert(orderHot)
				log.Errorf("%s归集错误,获取发送交易异常:%s", c.coinName, err.Error())
				// 写入热钱包表，创建失败
				log.Errorf(err.Error())
				continue
			}
			orderHot.TxId = txid
			orderHot.Status = int(status.BroadcastStatus)
			//保存热表
			err = dao.FcOrderHotInsert(orderHot)
			if err != nil {
				err = fmt.Errorf("[%s]归集保存订单[%s]数据异常:[%s]", c.coinName, orderHot.OuterOrderNo, err.Error())
				//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
				log.Error(err.Error())
			}
		}
	}
	return nil
}

func (c *DotRecycleService) walletServerCreateHot(orderReq *transfer.DotOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[c.coinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", c.coinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, c.coinName), cfg.User, cfg.Password, orderReq)
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
