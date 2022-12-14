package transfer

import (
	"bytes"
	"crypto/sha512"
	"encoding/base32"
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
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type AlgoTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewAlgoTransferService() service.TransferService {
	return &AlgoTransferService{
		CoinName: "algo",
		Lock:     &sync.Mutex{},
	}
}

func (s *AlgoTransferService) VaildAddr(address string) error {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(address)
	if err != nil {
		return fmt.Errorf("base32 decode error: %v", err)
	}
	// Ensure the decoded address is the correct length
	if len(decoded) != sha512.Size256+4 {
		err = fmt.Errorf("decode address length is not equal 36: %d", len(decoded))
		return err
	}
	// Split into address + checksum
	addressBytes := decoded[:sha512.Size256]
	checksumBytes := decoded[sha512.Size256:]

	// Compute the expected checksum
	checksumHash := sha512.Sum512_256(addressBytes)
	expectedChecksumBytes := checksumHash[sha512.Size256-4:]

	// Check the checksum
	if !bytes.Equal(expectedChecksumBytes, checksumBytes) {
		err = fmt.Errorf("address checksum is incorrect, did you copy the address correctly?,address: %s", address)
		return err
	}

	return nil
}

func (s *AlgoTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.AlgoOrderRequest
		amount     decimal.Decimal //????????????
		createData []byte          //??????????????????
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
		return "", fmt.Errorf("??????????????????")
	}
	orderReq, err = s.buildOrderHot(ta, int32(coinSet.Decimal))
	if err != nil {
		log.Errorf("???????????????id???%d,????????????:%s", ta.Id, err.Error())
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
		Amount:       amount.IntPart(), //????????????
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
		log.Errorf("???????????????id???%d,????????????????????????:%s", ta.Id, err.Error())
		// ?????????????????????????????????
		return "", err
	}
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	// ????????????
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("????????????[%s]????????????:[%s]", orderHot.OuterOrderNo, err.Error())
		// ?????????????????????,????????????????????????txid???????????????????????????????????????
		log.Error(err.Error())
		// ???????????????
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}

func (s *AlgoTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================????????????=================
//????????????????????????
func (s *AlgoTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {
	return nil
}

func (s *AlgoTransferService) walletServerCreateHot(orderReq *transfer.AlgoOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s ?????????????????? :%s", s.CoinName, string(dd))
	log.Infof("%s ?????????????????? :%s", s.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order??? ???????????????????????????outOrderId???%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Txid == "" {
		return "", fmt.Errorf("order??? ?????????????????????????????????,????????????????????????outOrderId???%s???err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Txid, nil
}
func (s *AlgoTransferService) buildOrderHot(ta *entity.FcTransfersApply, coinDecimal int32) (*transfer.AlgoOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	// ??????from???????????????
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
	//??????????????????????????????
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("????????????ID???%d?????????????????????%s,???????????????????????????", ta.Id, ta.OutOrderid)
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
	fee := decimal.NewFromFloat(0.001).Shift(-coinDecimal)
	if ta.Fee != "" {
		fee, _ = decimal.NewFromString(ta.Fee)
	}
	fromAmount := toAmount.Add(fee)
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount > ? and forzen_amount = 0", coinType, fromAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s ??????????????????????????????????????????1.1 \n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))]

	orderReq := &transfer.AlgoOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Value = toAmount.Shift(coinDecimal).String()
	orderReq.Fee = fee.Shift(coinDecimal).String()
	if ta.Eostoken != "" {
		orderReq.Assert = ta.Eostoken
	} else {
		orderReq.Assert = "0"
	}
	return orderReq, nil
}

//???????????? ??????eth??????
