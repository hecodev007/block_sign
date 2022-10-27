package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	addr "github.com/filecoin-project/go-address"
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
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type FilTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewFilTransferService() service.TransferService {
	return &FilTransferService{
		CoinName: "fil",
		Lock:     &sync.Mutex{},
	}
}

func (s *FilTransferService) VaildAddr(address string) error {
	if len(address) == 0 {
		return errors.New("address is null")
	}
	if len(address) > addr.MaxAddressStringLength || len(address) < 3 {
		return addr.ErrInvalidLength
	}
	//if conf.Cfg.Mode == "prod" {
	//	if strings.HasPrefix(address,addr.MainnetPrefix)   {
	//		return addr.ErrUnknownNetwork
	//	}
	//} else {
	//	if string(address[0]) != addr.MainnetPrefix || string(address[0]) != addr.TestnetPrefix {
	//		return addr.ErrUnknownNetwork
	//	}
	//}
	if !strings.HasPrefix(address, addr.MainnetPrefix) {
		return addr.ErrUnknownNetwork
	}
	var protocol addr.Protocol
	switch address[1] {
	case '0':
		protocol = addr.ID
		return errors.New("暂不支持addr.ID格式地址")
	case '1':
		protocol = addr.SECP256K1
	case '2':
		protocol = addr.Actor
		//return errors.New("暂不支持addr.Actor格式地址")
	case '3':
		protocol = addr.BLS
		//return errors.New("暂不支持addr.BLS格式地址")
	default:
		return addr.ErrUnknownProtocol
	}
	raw := address[2:]
	payloadcksm, err := addr.AddressEncoding.WithPadding(-1).DecodeString(raw)
	if err != nil {
		return err
	}
	payload := payloadcksm[:len(payloadcksm)-addr.ChecksumHashLength]
	cksm := payloadcksm[len(payloadcksm)-addr.ChecksumHashLength:]
	if protocol == addr.SECP256K1 || protocol == addr.Actor {
		if len(payload) != 20 {
			return addr.ErrInvalidPayload
		}
	}
	if !addr.ValidateChecksum(append([]byte{protocol}, payload...), cksm) {
		return addr.ErrInvalidChecksum
	}
	return nil
}

func (s *FilTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.FilOrderRequest
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
	orderReq, err = s.buildOrderHot(ta, int32(coinSet.Decimal))
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount, _ = decimal.NewFromString(orderReq.Amount)

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
		Amount:       amount.IntPart(), //转换整型
		Quantity:     amount.String(),
		Decimal:      int64(coinSet.Decimal),
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
	// 保存热表
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

func (s *FilTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================私有方法=================
//创建交易接口参数
func (s *FilTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {
	return nil
}

func (s *FilTransferService) walletServerCreateHot(orderReq *transfer.FilOrderRequest) (string, error) {
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
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Txid, nil
}
func (s *FilTransferService) buildOrderHot(ta *entity.FcTransfersApply, coinDecimal int32) (*transfer.FilOrderRequest, error) {
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
	//
	fromAmount := toAmount.Add(decimal.NewFromFloat(0.0001))
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount > ? and forzen_amount = 0", coinType, fromAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址，预扣手续费0.0001 \n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))]

	orderReq := &transfer.FilOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(coinDecimal).String()
	orderReq.GasLimit = 4000000
	orderReq.GasPremium = 5
	orderReq.GasFeeCap = 0
	return orderReq, nil
}
