package transfer

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"strconv"

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
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"

	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type SolTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewSolTransferService() service.TransferService {
	return &SolTransferService{
		CoinName: "sol",
		Lock:     &sync.Mutex{},
	}
}

func (s *SolTransferService) VaildAddr(address string) error {
	return s.validSolAddress(address)
}

func (s *SolTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.SolOrderRequest
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

func (s *SolTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================私有方法=================
//创建交易接口参数
func (s *SolTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
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

func (s *SolTransferService) walletServerCreateHot(orderReq *transfer.SolOrderRequest) (string, error) {
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
	result := transfer.DecodeSolTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result["data"].(string), nil
}
func (s *SolTransferService) buildOrderHot(dec int32, ta *entity.FcTransfersApply) (*transfer.SolOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)

	SolnewCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": ta.CoinName}))
	if err != nil {
		return nil, err
	}
	if len(SolnewCoins) == 0 {
		return nil, fmt.Errorf("do not find %s coin", ta.CoinName)
	}
	pid := SolnewCoins[0].Id
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"pid": pid}))
	if err != nil {
		return nil, err
	}
	solToken := ""
	for _, a := range coins {
		if strings.ToLower(a.Token) == strings.ToLower(ta.Eostoken) {
			solToken = a.Token
		}
	}

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
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))]

	//填充参数
	orderReq := &transfer.SolOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.OrderId = ta.OrderId
	orderReq.MchId = strconv.Itoa(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(dec).String()
	if ta.Eoskey != "" {
		orderReq.ContractAddress = ta.Eostoken
	}
	if solToken != "" {
		orderReq.OriginalContractAddress = solToken
	}
	return orderReq, nil
}

func (s *SolTransferService) validSolAddress(address string) error {
	data := base58.Decode(address)
	if len(data) == 0 {
		return errors.New("bas58 decode data error")
	}
	if len(data) != 32 {
		return errors.New("public key length is not equal 32")
	}
	return nil
}
