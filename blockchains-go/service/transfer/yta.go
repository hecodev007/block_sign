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
	"strconv"
	"strings"
	"sync"
	"time"
)

type YtaTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewYtaTransferService() service.TransferService {
	return &YtaTransferService{
		CoinName: "yta",
		Lock:     &sync.Mutex{},
	}
}
func (s *YtaTransferService) VaildAddr(address string) error {
	if len(address) == 0 || len(address) > 12 {
		return fmt.Errorf("address length is not correct,length=%d", len(address))
	}
	u, err := s.stringToName(address)
	if err != nil {
		return fmt.Errorf("yta地址转换为uint64错误: %v", err)
	}
	address2 := s.nameToString(u)
	if address != address2 {
		return fmt.Errorf("yta uint64转换为地址错误，原地址为： %s，转换后地址为： %s ", address, address2)
	}
	return nil
}
func (s *YtaTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.YtaOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinSet = global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, amount, err = s.buildOrderHot(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	orderReq.MchId = int64(mch.Id)
	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.Data.FromAddress,
		ToAddress:    orderReq.Data.ToAddress,
		Token:        orderReq.Data.Token,
		Amount:       amount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:     amount.String(),
		Memo:         orderReq.Data.Memo,
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
		//写入热钱包表，创建失败
		return "", err
	}
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	//保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		//发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}
func (s *YtaTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	panic("implement me")
}

//私有方法 构建wax订单
func (s *YtaTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.YtaOrderRequest, decimal.Decimal, error) {
	var (
		fromAddr string
		pubKey   string
		toAddr   string
		toAmount decimal.Decimal
	)
	type key struct {
		Key string `json:"key"`
	}
	pubKeyStruct := new(key)
	coinName := ta.CoinName
	if ta.Eoskey != "" {
		coinName = ta.Eoskey
	}
	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, decimal.Zero, fmt.Errorf("yta 缺少币种设置：%s", coinName)
	}
	// 查找from地址和金额
	fromAddrs, err := dao.FcGenerateAddressListFindAddressesData(int(address.AddressTypeCold), int(address.AddressStatusAlloc), ta.AppId, ta.CoinName)
	if err != nil {
		return nil, decimal.Zero, err
	}
	if len(fromAddrs) == 0 {
		return nil, decimal.Zero, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0].Address
	if fromAddrs[0].Json == "" {
		return nil, decimal.Zero, errors.New("缺少设置出账地址公钥失败")
	}
	json.Unmarshal([]byte(fromAddrs[0].Json), pubKeyStruct)

	pubKey = pubKeyStruct.Key
	if pubKey == "" {
		return nil, decimal.Zero, errors.New("查询出账地址公钥失败")
	}
	//查询出账地址和金额
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return nil, decimal.Zero, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, decimal.Zero, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	//toAmount = toAddrs[0].ToAmount
	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
	if toAmount.IsZero() {
		return nil, decimal.Zero, errors.New("ysr toAmount  is zero")
	}
	var orderData transfer.YtaOrderData
	toAmount2, _ := toAmount.Float64()
	orderData.Quantity = fmt.Sprintf("%."+strconv.Itoa(coinSet.Decimal)+"f"+" %s", toAmount2, strings.ToUpper(coinName))
	orderData.Memo = ta.Memo
	orderData.Token = coinSet.Token
	orderData.FromAddress = fromAddr
	orderData.ToAddress = toAddr
	orderData.SignPubkey = pubKey

	orderReq := new(transfer.YtaOrderRequest)
	orderReq.CoinName = coinName
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.Data = orderData
	return orderReq, toAmount, nil
}

//创建交易接口参数
func (s *YtaTransferService) walletServerCreateHot(orderReq *transfer.YtaOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	url := fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName)
	log.Info("yta url:", url)
	data, err := util.PostJsonByAuth(url, cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result.Data.(string), nil
}

//验证地址
/*
https://github.com/eoscanada/eos-go/name.go
*/

func (s *YtaTransferService) stringToName(ss string) (val uint64, err error) {
	// ported from the eosio codebase, libraries/chain/include/eosio/chain/name.hpp
	var i uint32
	sLen := uint32(len(ss))
	for ; i <= 12; i++ {
		var c uint64
		if i < sLen {
			c = uint64(s.charToSymbol(ss[i]))
		}
		if i < 12 {
			c &= 0x1f
			c <<= 64 - 5*(i+1)
		} else {
			c &= 0x0f
		}

		val |= c
	}

	return
}

func (s *YtaTransferService) charToSymbol(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 6
	}
	if c >= '1' && c <= '5' {
		return c - '1' + 1
	}
	return 0
}

func (s *YtaTransferService) nameToString(in uint64) string {
	// Some particularly used name are pre-cached, so we avoid the transformation altogether, and reduce memory usage
	// ported from libraries/chain/name.cpp in eosio
	a := []byte{'.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}

	tmp := in
	i := uint32(0)
	for ; i <= 12; i++ {
		bit := 0x1f
		if i == 0 {
			bit = 0x0f
		}
		c := base32Alphabet[tmp&uint64(bit)]
		a[12-i] = c

		shift := uint(5)
		if i == 0 {
			shift = 4
		}

		tmp >>= shift
	}

	// We had a call to `strings.TrimRight` before, but that was causing lots of
	// allocation and lost CPU cycles. We now have our own cutting method that
	// improves performance a lot.
	return s.trimRightDots(a)
}

var base32Alphabet = []byte(".12345abcdefghijklmnopqrstuvwxyz")

func (s *YtaTransferService) trimRightDots(bytes []byte) string {
	trimUpTo := -1
	for i := 12; i >= 0; i-- {
		if bytes[i] == '.' {
			trimUpTo = i
		} else {
			break
		}
	}

	if trimUpTo == -1 {
		return string(bytes)
	}

	return string(bytes[0:trimUpTo])
}
