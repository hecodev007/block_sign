package transfer

import (
	"bytes"
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
	"github.com/fioprotocol/fio-go/eos/btcsuite/btcutil/base58"
	"github.com/fioprotocol/fio-go/eos/ecc"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	_ "github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"golang.org/x/crypto/ripemd160"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type FioTransferService struct {
	CoinName string
	Lock     *sync.Mutex
	limitMap sync.Map
}

func NewFioTransferService() service.TransferService {
	return &FioTransferService{
		CoinName: "fio",
		Lock:     &sync.Mutex{},
		limitMap: sync.Map{},
		//lastFromAddress: "",
	}
}

func (s *FioTransferService) VaildAddr(address string) error {
	return validAddress(address)
}

func (s *FioTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.FioOrderRequest
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
	// feeInt64, _ = decimal.NewFromString(result.Fee)
	// feeInt64 = feeInt64.Shift(int32(coinSet.Decimal))
	// orderHot.Fee = feeInt64.IntPart()
	orderHot.TxId = txid
	s.limitMap.Store(orderReq.FromAddress, time.Now().Unix())
	// orderHot.MemoEncrypt = result.Memo
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

func (s *FioTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	// orderReq, err := s.buildOrder(ta)
	// if err != nil {
	// 	log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
	// 	return err
	// }
	// err = s.walletServerCreate(orderReq)
	// if err != nil {
	// 	log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
	// 	return err
	// }
	// return nil
	return errors.New("do not support cold transfer")
}

//=================私有方法=================
//创建交易接口参数
func (s *FioTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {
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

func (s *FioTransferService) walletServerCreateHot(orderReq *transfer.FioOrderRequest) (string, error) {
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
	result := transfer.DecodeFioTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result["data"].(string), nil
}
func (s *FioTransferService) buildOrderHot(dec int32, ta *entity.FcTransfersApply) (*transfer.FioOrderRequest, error) {
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
	//fromAddr = fromAddrs[rand.Intn(len(fromAddrs))]
	fromAddr, err = s.findFromAddress(fromAddrs)
	if err != nil {
		return nil, fmt.Errorf("from地址处于冻结时间，冻结时间为2分钟")
	}
	//填充参数
	orderReq := &transfer.FioOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	// orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(dec).String()

	return orderReq, nil
}

func validAddress(address string) error {
	if !strings.HasPrefix(address, "FIO") {
		return errors.New("dont have 'FIO' prefix")
	}
	data := base58.Decode(address[3:])
	if len(data) != 37 {
		return errors.New("bas58 decode data length is not equal 37")
	}
	checksum1 := data[33:]
	ck2 := ripemd160checksum(data[:33], ecc.CurveK1)
	checksum2 := ck2[:4]
	if bytes.Compare(checksum1, checksum2) != 0 {
		return errors.New("checksum is not equal")
	}
	return nil
}

func ripemd160checksum(in []byte, curve ecc.CurveID) []byte {
	h := ripemd160.New()
	_, _ = h.Write(in) // this implementation has no error path

	if curve != ecc.CurveK1 {
		_, _ = h.Write([]byte(curve.String()))
	}

	sum := h.Sum(nil)
	return sum[:4]
}

func (s *FioTransferService) findFromAddress(addresses []string) (string, error) {
	if len(addresses) == 0 {
		return "", errors.New("addresses length is zero")
	}
	for _, addr := range addresses {
		v, ok := s.limitMap.Load(addr)
		if !ok {
			return addr, nil
		}
		lastTxTime := v.(int64)
		//冻结2分钟
		frzenzTime := addTimeSecond(lastTxTime, 2)
		if frzenzTime > time.Now().Unix() {
			continue
		}
		return addr, nil
	}
	return "", errors.New("do not find any from address")
}

func addTimeSecond(now, minute int64) int64 {
	return time.Unix(now, 0).Add(time.Minute * time.Duration(minute)).Unix()
}
