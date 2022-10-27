package transfer

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
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

/*
func: stx新链
author: flynn
date: 2020-02-01
*/

type StxNewTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewStxNewTransferService() service.TransferService {
	return &StxNewTransferService{
		CoinName: "stx",
		Lock:     &sync.Mutex{},
	}
}

func (s *StxNewTransferService) VaildAddr(address string) error {
	//print(s.getBalance("SP2KEVFAFDQ83ZGPJSEP6H9A856D5HR6ZR02WR3V7", "nyc"))

	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return fmt.Errorf("valid address: don't find %s config", s.CoinName)
	}
	params := make(map[string]interface{})
	params["address"] = address
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/validAddress", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return fmt.Errorf("%s rpc valid address error: %v", s.CoinName, err)
	}
	//print(s.getBalance("SP2X88547NDNQ4WZFV90G0BVCRB0AFZKNDPRR0DJC", "nyc"))
	return transfer.DecodeValidAddressResp(data)
}

func (s *StxNewTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.StxNewOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err = s.buildOrderHot(int32(coinSet.Decimal), ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount, _ = decimal.NewFromString(orderReq.Value)

	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.FromAddress,
		ToAddress:    orderReq.ToAddress,
		Amount:       amount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:     amount.String(),
		Decimal:      int64(global.CoinDecimal[ta.CoinName].Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
	}

	orderHot.TxId = txid

	orderHot.Status = int(status.BroadcastStatus)
	//保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}

func (s *StxNewTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================私有方法=================

func (s *StxNewTransferService) walletServerCreateHot(orderReq *transfer.StxNewOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if err != nil || result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Txid == "" {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result.Txid, nil
}
func (s *StxNewTransferService) buildOrderHot(dec int32, ta *entity.FcTransfersApply) (*transfer.StxNewOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, err = decimal.NewFromString(toAddrs[0].ToAmount)
	if err != nil {
		return nil, err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	//
	fromAddr = fromAddrs[0]
	// 判断出账地址金额是否足够
	fromAmount, err := s.getBalance(fromAddr, ta.Eoskey)
	if err != nil {
		return nil, err
	}
	fa, _ := decimal.NewFromString(fromAmount)
	if fa.LessThan(toAmount) {
		return nil, fmt.Errorf("冷地址链上金额不足出账，coldChainAmt=[%s],transferAmt=[%s] Eoskey[%s]",
			fa.String(), toAmount.String(), ta.Eoskey)
	}

	//填充参数
	orderReq := &transfer.StxNewOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	//orderReq. = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = strings.ToUpper(ta.CoinName)

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Value = toAmount.String()
	orderReq.Memo = ta.Memo

	//默认值
	orderReq.Nonce = 0
	orderReq.Fee = "0"

	if ta.Eostoken != "" {
		coin := global.CoinDecimal[ta.Eoskey]
		if coin == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", ta.Eoskey)
		}
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.ContractAddress = coin.Token
		orderReq.Token = coin.Name
		orderReq.Value = toAmount.String()
	}
	fmt.Printf("buildOrderHot orderReq: %v\n", orderReq)
	return orderReq, nil
}

func (s *StxNewTransferService) getBalance(address, tokenName string) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("get balance: don't find %s config", s.CoinName)
	}
	params := make(map[string]interface{})
	params["address"] = address
	params["token"] = tokenName
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/getBalance", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return "", fmt.Errorf("%s rpc get balance error: %v", s.CoinName, err)
	}
	resp, err := transfer.DecodeGetBalanceResp(data)
	if err != nil {
		return "", err
	}
	balance, isOk := resp.Data.(string)
	if !isOk {
		return "", fmt.Errorf("get balance resp data is not string: %v", resp.Data)
	}
	return balance, nil
}
