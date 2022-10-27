package transfer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
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

type HntTransferService struct {
	CoinName     string
	Lock         *sync.Mutex
	CurFromIndex int
}

func NewHntTransferService() service.TransferService {
	return &HntTransferService{
		CoinName:     "hnt",
		Lock:         &sync.Mutex{},
		CurFromIndex: 0,
	}
}

func DoubleSha256(data []byte) []byte {
	d := sha256.Sum256(data)
	dd := sha256.Sum256(d[:])
	return dd[:]
}
func (s *HntTransferService) VaildAddr(address string) error {

	data := base58.Decode(address)
	if data == nil || len(data) == 0 {
		return errors.New("base58 decode address error")
	}

	if data[0] != byte(0) {
		return errors.New("invalid version")
	}
	if data[1] > byte(1) {
		return fmt.Errorf("invalid curve version,[0 is NIST p256,1 is ed25519],curve=%d", data[1])
	}
	checksum1 := data[len(data)-4:]
	ck := DoubleSha256(data[:len(data)-4])
	checksum2 := ck[:4]
	if !bytes.Equal(checksum1, checksum2) {
		return errors.New("invalid checksum")
	}
	return nil
}

func (s *HntTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.HntOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
		txFeeFolat string
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
	txid, txFeeFolat, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		if s.CurFromIndex > 0 {
			s.CurFromIndex--
		}
		log.Infof("hnt 签名失败 索引回滚至 %d", s.CurFromIndex)
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
	}
	feeFloat, _ := decimal.NewFromString(txFeeFolat)
	orderHot.TxId = txid
	orderHot.Fee = feeFloat.Shift(8).IntPart()
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

func (s *HntTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================私有方法=================
//创建交易接口参数
func (s *HntTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {

	return nil
}

func (s *HntTransferService) walletServerCreateHot(orderReq *transfer.HntOrderRequest) (txid, feeFloat string, err error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	thr, err1 := transfer.DecodeHntTransferHotResp(data)
	if err1 != nil {
		return "", "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Data.Txid == "" {
		return "", "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.Txid, thr.Data.FeeFloat, nil
}
func (s *HntTransferService) buildOrderHot(ta *entity.FcTransfersApply, coinDecimal int32) (*transfer.HntOrderRequest, error) {
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
	//fromAmount := toAmount.Add(decimal.NewFromFloat(0.001))
	toAmountAddFee := toAmount.Add(decimal.NewFromFloat(1))
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount > ? and forzen_amount = 0", coinType, toAmountAddFee).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址,预扣1手续费\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	idx := s.CurFromIndex % len(fromAddrs)
	fromAddr = fromAddrs[idx]
	log.Infof("hnt 使用第 %d 个索引地址 %s", idx, fromAddr)

	orderReq := &transfer.HntOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(coinDecimal).String()

	s.CurFromIndex++
	return orderReq, nil
}

//私有方法 构建eth订单
