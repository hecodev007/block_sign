package transfer

import (
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
	"github.com/ethereum/go-ethereum/common"
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
	"math/rand"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

const BscDecimal = 18

type BscTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewBscTransferService() service.TransferService {
	return &BscTransferService{
		CoinName: "bsc",
		Lock:     &sync.Mutex{},
	}
}

func (s *BscTransferService) VaildAddr(address string) error {

	if !s.has0xPrefix(address) {
		return errors.New("address must start with 0x")
	}
	if !common.IsHexAddress(address) {
		return errors.New("valid bsc address error")
	}
	return nil

	//data, err := util.Get(fmt.Sprintf("%s/%s/validaddress?address=%s", conf.Cfg.Walletserver.Url, s.CoinName, address))
	//if err != nil {
	//	return err
	//}
	//log.Infof("vaild address :%s", string(data))
	//result := transfer.DecodeWalletServerRespOrder(data)
	//if result == nil {
	//	return fmt.Errorf("请求验证接口失败，address：%s", address)
	//}
	//if result.Code != 0 || result.Data == nil {
	//	log.Error(result)
	//	if result.Message == "contractAddrNotAllow" {
	//		return errors.New(result.Message)
	//	}
	//
	//	return fmt.Errorf("请求验证接口失败,服务器返回异常，address：%s", address)
	//}
	//res, ok := result.Data.(bool)
	//if !ok {
	//	return fmt.Errorf("data type err")
	//}
	//if !res {
	//	return fmt.Errorf("address is avalid")
	//}
	//return nil
}

// has0xPrefix validates str begins with '0x' or '0X'.
func (s *BscTransferService) has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func (s *BscTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.BscOrderRequest
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
	orderReq, err = s.buildOrderHot(ta)
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
	if orderReq.ContractAddress != "" {
		orderReq.Token = orderReq.ContractAddress
	}
	txid, err = s.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())

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

func (s *BscTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := s.buildOrderCold(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return err
	}
	err = s.walletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		return err
	}
	return nil
}

//=================私有方法=================
//创建交易接口参数
func (s *BscTransferService) walletServerCreate(orderReq *transfer.BscOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		if strings.Contains(err.Error(), "main net fee too high") {
			return errors.New("BSC主网手续费过高，请值班在留意主网手续费并重推")
		}
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

//私有方法 构建bsc订单
func (s *BscTransferService) buildOrderCold(ta *entity.FcTransfersApply) (*transfer.BscOrderRequest, error) {
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
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查询出账地址和金额遇到错误, err: %s", ta.Id, ta.OutOrderid, err.Error())
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

	searchAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	if strings.ToLower(coinType) == "eth" {
		searchAmount = searchAmount.Add(decimal.NewFromFloat(0.1))
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, searchAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		dao.ReportColdAddrBalanceNotEnough(int64(ta.AppId), ta.OutOrderid, "bsc", coinType, searchAmount.String())
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	//fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	fromAddr = fromAddrs[0] //规则修改：取第一个
	//填充参数
	orderReq := &transfer.BscOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
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
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.Token = ta.Eostoken
		orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).String()
	} else {
		orderReq.Amount = toAmount.Shift(int32(BscDecimal)).String()
	}
	return orderReq, nil
}

func (s *BscTransferService) walletServerCreateHot(orderReq *transfer.BscOrderRequest) (string, error) {
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
	result := transfer.DecodeBscTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result["data"].(string), nil
}

func (s *BscTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.BscOrderRequest, error) {
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
		dao.ReportColdAddrBalanceNotEnough(int64(ta.AppId), ta.OutOrderid, "bsc", coinType, toAddrs[0].ToAmount)
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]
	//填充参数
	orderReq := &transfer.BscOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
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
		orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).String()
	} else {
		orderReq.Amount = toAmount.Shift(int32(BscDecimal)).String()
	}
	return orderReq, nil
}

func (s *BscTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.EthOrderRequest, error) {
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
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	//填充参数
	orderReq := &transfer.EthOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
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
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.Token = ta.Eostoken
		orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).String()
	} else {
		orderReq.Amount = toAmount.Shift(int32(BscDecimal)).String()
	}
	return orderReq, nil
}
