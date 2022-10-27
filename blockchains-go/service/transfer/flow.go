package transfer

import (
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
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
	"github.com/onflow/flow-go-sdk"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

const FlowDecimal = 8

type FlowTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewFlowTransferService() service.TransferService {
	return &FlowTransferService{
		CoinName: "flow",
		Lock:     &sync.Mutex{},
	}
}

func (s *FlowTransferService) VaildAddr(address string) error {
	if flowAddrVerify(address) {
		return nil
	}
	return errors.New("地址格式错误")
}

func flowAddrVerify(addr string) bool {
	fAddr := flow.HexToAddress(addr)
	if fAddr.IsValid(flow.Mainnet) {
		return true
	}
	return false
}

func (s *FlowTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.FlowOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	log.Infof("TransferHot执行001，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinSet = global.CoinDecimal[ta.CoinName]

	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	log.Infof("TransferHot执行002，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
	orderReq, err = s.buildOrderHot(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount = decimal.NewFromInt(orderReq.Amount)
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
	if orderReq.ContractAddress != "" {
		orderReq.Token = orderReq.ContractAddress
	}

	log.Infof("TransferHot执行003，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
	txid, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		log.Infof("TransferHot执行004，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())

		return "", err
	}

	orderHot.TxId = txid

	orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	log.Infof("TransferHot执行005，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	log.Infof("TransferHot执行006，订单：%s,时间：", ta.OutOrderid, util.GetChinaTimeNowFormat())
	return txid, nil
}

func (s *FlowTransferService) TransferCold(ta *entity.FcTransfersApply) error {
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
////创建交易接口参数
//func (s *FlowTransferService) walletServerCreate(orderReq *transfer.EthOrderRequest) error {
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
//	if err != nil {
//		return err
//	}
//	dd, _ := json.Marshal(orderReq)
//	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
//	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
//	result := transfer.DecodeWalletServerRespOrder(data)
//	if result == nil {
//		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
//	}
//	if result.Code != 0 || result.Data == nil {
//		log.Error(result)
//		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
//	}
//	return nil
//}

func (s *FlowTransferService) walletServerCreateHot(orderReq *transfer.FlowOrderRequest) (string, error) {
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
	result := transfer.DecodeFlowTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	txId := result["data"].(string)

	return txId, nil
}
func (s *FlowTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.FlowOrderRequest, error) {
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
	fromAddr = fromAddrs[0]
	//填充参数
	orderReq := &transfer.FlowOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = ta.Applicant
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr

	if ta.Eostoken != "" {
		coin := global.CoinDecimal[ta.Eoskey]
		if coin == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", ta.Eoskey)
		}
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.ContractAddress = ta.Eostoken
		orderReq.Token = ta.Eoskey
		orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).IntPart()
	} else {
		orderReq.Amount = toAmount.Shift(int32(FlowDecimal)).IntPart()
	}
	return orderReq, nil
}
