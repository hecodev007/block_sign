package transfer

import (
	"bytes"
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/group-coldwallet/blockchains-go/conf"
	_ "github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"golang.org/x/crypto/sha3"
	"strings"
	"sync"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

const NasDecimal = 18

type NasTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewNasTransferService() service.TransferService {
	return &NasTransferService{
		CoinName: "nas",
		Lock:     &sync.Mutex{},
	}
}

func (s *NasTransferService) VaildAddr(address string) error {
	return s.validAddress(address)
}

func (s *NasTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		//mch        *entity.FcMch
		coinSet  *entity.FcCoinSet
		orderReq *transfer.NasOrderRequest
		//amount     decimal.Decimal //发送金额
		//createData []byte          //构造交易信息
	)
	//mch, err = dao.FcMchFindById(ta.AppId)
	//if err != nil {
	//	return "", err
	//}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err = s.buildOrderHot(ta, coinSet)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	//amount, _ = decimal.NewFromString(orderReq.Amount)

	//createData, _ = json.Marshal(orderReq)
	//orderHot := &entity.FcOrderHot{
	//	ApplyId:      ta.Id,
	//	ApplyCoinId:  coinSet.Id,
	//	OuterOrderNo: ta.OutOrderid,
	//	OrderNo:      ta.OrderId,
	//	MchName:      mch.Platform,
	//	CoinName:     ta.CoinName,
	//	FromAddress:  orderReq.FromAddress,
	//	ToAddress:    orderReq.ToAddress,
	//	Amount:       amount.IntPart(), //转换整型
	//	Quantity:     amount.String(),
	//	Decimal:      int64(coinSet.Decimal),
	//	CreateData:   string(createData),
	//	Status:       int(status.UnknowErrorStatus),
	//	CreateAt:     time.Now().Unix(),
	//	UpdateAt:     time.Now().Unix(),
	//}
	txid, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		//orderHot.Status = int(status.BroadcastErrorStatus)
		//orderHot.ErrorMsg = err.Error()
		//dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
	}
	//orderHot.TxId = txid
	//orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	//err = dao.FcOrderHotInsert(orderHot)
	//if err != nil {
	//	//err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
	//	// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
	//	log.Error(err.Error())
	//	// 发送给钉钉
	//	dingding.ErrTransferDingBot.NotifyStr(err.Error())
	//}
	return txid, nil
}

func (s *NasTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support it")
}

//=================私有方法=================

func (s *NasTransferService) buildOrderHot(ta *entity.FcTransfersApply, coin *entity.FcCoinSet) (*transfer.NasOrderRequest, error) {
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
	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}

	//fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	fee := decimal.NewFromFloat(0.02) //手续费
	for _, from := range fromAddrs {
		fromAmount, _ := decimal.NewFromString(from.Amount)
		fromAmount = fromAmount.Sub(fee)
		if fromAmount.GreaterThanOrEqual(toAmount) {
			log.Infof(fromAmount.String())
			log.Infof(toAmount.String())
			fromAddr = from.Address
			break
		}
	}
	if fromAddr == "" {
		return nil, errors.New("from address sub fee is not enougth transfer")
	}
	//填充参数
	orderReq := &transfer.NasOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.ApplyCoinId = int64(coin.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr

	if ta.Eostoken != "" { //如果是代币转账
		coin := global.CoinDecimal[ta.Eoskey]
		if coin == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", ta.Eoskey)
		}
		log.Infof("合约地址： %s", coin.Token)
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		//	不能使用ta.EosToken ,ta.EosToken 已经变成小写了，导致验证地址不合法
		log.Infof("使用的合约地址： %s", coin.Token)
		orderReq.Token = coin.Token
		orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).String()
	} else {
		orderReq.Amount = toAmount.Shift(int32(NasDecimal)).String()
	}
	return orderReq, nil
}

func (s *NasTransferService) walletServerCreateHot(orderReq *transfer.NasOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/nebulas/Transfer", cfg.Url), cfg.User, cfg.Password, orderReq)
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
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.(string), nil
}

func (s *NasTransferService) validAddress(address string) error {
	if len(address) == 0 {
		return errors.New("address is null")
	}
	data := base58.Decode(address)
	if len(data) == 0 {
		return errors.New("base58 decode error")
	}
	// address not equal to 26
	if len(data) != 26 {
		return errors.New("nas address length is not equal 26")
	}
	// check if address start with AddressPrefix = 25
	if data[0] != 25 {
		return errors.New("nas address prefix bytes is not equal 25")
	}
	// check if address type is NormalType = 87 or ContractType = 88
	if data[1] == 87 || data[1] == 88 {
		//check checksum
		content := data[:22]
		checksum := data[22:]
		hash := sha3.New256()
		hash.Write(content)
		c := hash.Sum(nil)
		cSum := c[:4]
		if bytes.Compare(checksum, cSum) != 0 {
			return errors.New("nas address check sum failed")
		}
	} else {
		return fmt.Errorf("nas address is type is not equal 87 or 88,type=[%d]", data[1])
	}
	return nil
}
