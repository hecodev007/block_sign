package recycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"time"
	"xorm.io/builder"
)

type CkbRecycleService struct {
	coinName string
}

func NewCkbRecycleService() *CkbRecycleService {
	return &CkbRecycleService{coinName: "ckb"}
}

func (c *CkbRecycleService) RecycleCoin(mchInfo *entity.FcMch) error {
	//获取币种的配置
	ckbCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": c.coinName}))
	if err != nil {
		return err
	}
	if len(ckbCoins) == 0 {
		return errors.New("do not find ckb coin")
	}
	//只有一个coin
	coin := ckbCoins[0]
	ta := &entity.FcTransfersApply{
		Username:   "Robot",
		Department: "blockchains-go",
		Applicant:  mchInfo.Platform,
		OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Operator:   "Robot",
		CoinName:   c.coinName,
		Type:       "gj",
		Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
		Status:     int(entity.ApplyStatus_Merge),
		Createtime: time.Now().Unix(),
		Lastmodify: time.Now(),
		Fee:        decimal.NewFromFloat(0.001).String(),
		AppId:      mchInfo.Id,
		Source:     1,
	}
	orderReq, err := c.buildOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return err
	}

	coldAddress, err1 := dao.FcGenerateAddressListFindAddressesData(1, 2, ta.AppId, c.coinName)
	if err1 != nil {
		log.Errorf("下单表订单id：%d,填充地址记录异常:%s", ta.Id, err1.Error())
		return err1
	}
	if len(coldAddress) == 0 {
		log.Errorf("下单表订单id：%d,查找冷地址记录异常:%s", ta.Id, err1.Error())
		return err1
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     coldAddress[0].Address,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  ta.Lastmodify,
	})
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     coldAddress[0].Address,
		AddressFlag: "from",
		Status:      0,
		Lastmodify:  ta.Lastmodify,
	})

	appId, err := ta.TransactionAdd(applyAddresses)
	orderReq.ApplyId = appId
	orderReq.ApplyCoinId = int64(coin.Id)
	err = c.walletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		return err
	}
	return nil
}

//创建交易接口参数
func (s *CkbRecycleService) walletServerCreate(orderReq *transfer.CkbOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.coinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//私有方法 构建冷地址合并ckb订单
func (s *CkbRecycleService) buildOrder(ta *entity.FcTransfersApply) (*transfer.CkbOrderRequest, error) {
	//查找冷地址
	coldAddress, err1 := dao.FcGenerateAddressListFindAddressesData(1, 2, ta.AppId, s.coinName)
	if err1 != nil || len(coldAddress) == 0 {
		return nil, fmt.Errorf("无法获取币种%s冷地址", s.coinName)
	}
	from := coldAddress[0].Address
	to := from
	toAddrAmount := decimal.New(-8, 0)
	//	构建订单
	orderReq := &transfer.CkbOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     s.coinName,
			Worker:       service.GetWorker(s.coinName),
		},
	}
	var orderAddreses []map[string]interface{}
	fromOrder := map[string]interface{}{
		"dir":     0,
		"address": from,
	}
	toOrder := map[string]interface{}{
		"dir":      1,
		"address":  to,
		"quantity": toAddrAmount.String(),
	}
	changeOrder := map[string]interface{}{
		"dir":     2,
		"address": from,
	}
	orderAddreses = append(orderAddreses, fromOrder)
	orderAddreses = append(orderAddreses, toOrder)
	orderAddreses = append(orderAddreses, changeOrder)

	orderReq.OrderAddress = orderAddreses
	orderReq.FeeString = ta.Fee
	return orderReq, nil
}
