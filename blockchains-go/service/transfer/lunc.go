package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	"xorm.io/builder"
)

const LuncDecimal = 6

type LuncTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewLuncTransferService() service.TransferService {
	return &LuncTransferService{CoinName: "lunc",
		Lock: &sync.Mutex{}}
}

func (s *LuncTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
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
	orderReq, err := s.buildOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount = decimal.NewFromInt(orderReq.Data.Amount)
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
		Amount:       amount.IntPart(), //转换整型
		Quantity:     amount.String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err = s.walletServerCreate(orderReq)
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
func (s *LuncTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return fmt.Errorf("don't implement for lunc cold ")
}
func (s *LuncTransferService) VaildAddr(address string) error {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return fmt.Errorf("don't find %s config", s.CoinName)
	}
	params := make(map[string]string)
	params["address"] = address
	dd, _ := json.Marshal(params)

	log.Infof("url %v 验证地址%s 交易发送内容 :%s", fmt.Sprintf("%s/v1/%s/validAddress", cfg.Url, s.CoinName), s.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/validAddress", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return err
	}

	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	return transfer.DecodeValidAddressResp(data)
}

//创建交易接口参数
func (s *LuncTransferService) walletServerCreate(orderReq *transfer.LuncOrderRequest) (txid string, err error) {
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
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}

//======================私有方法==================
//私有方法 构建lunc订单
func (s *LuncTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.LuncOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	coinSet := global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少lunc币种信息")
	}

	// 查找from地址和金额
	fromAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	if len(fromAddrs) == 0 {
		return nil, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0] // 这里cocos只有一个
	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
	//填充参数
	orderReq := &transfer.LuncOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.ApplyCoinId = int64(coinSet.Id)
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Data = transfer.LuncPaymentRequest{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      toAmount.Shift(LuncDecimal).IntPart(), //转换成
		Memo:        ta.Memo,
		Token:       ta.Eostoken,
	}
	return orderReq, nil
}
