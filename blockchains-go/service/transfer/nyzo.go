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
	"strings"
	"time"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type NyzoTransferService struct {
	CoinName string
}

func NewNyzoTransferService() service.TransferService {
	return &NyzoTransferService{
		CoinName: "nyzo",
	}
}

func (s *NyzoTransferService) VaildAddr(address string) error {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return fmt.Errorf("valid address: don't find %s config", s.CoinName)
	}
	params := make(map[string]interface{})
	params["address"] = address
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/validAddress", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return fmt.Errorf("%s rpc valid address error: %v", s.CoinName, err)
	}
	return transfer.DecodeValidAddressResp(data)
}

func (s *NyzoTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.NyzoOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinType := s.CoinName

	if ta.Eostoken != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种(%s)信息", coinType)
	}
	orderReq, err = s.buildOrderHot(ta, coinSet)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
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
		Amount:       amount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
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

func (s *NyzoTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support cold transfer")
}

func (s *NyzoTransferService) walletServerCreateHot(orderReq *transfer.NyzoOrderRequest) (string, error) {
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
		return "", fmt.Errorf("order表 请求下单接口失败 ,err: %v", err1)
	}
	if thr.Code != 0 || thr.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，err:%s", string(data))
	}
	return thr.Txid, nil
}
func (s *NyzoTransferService) buildOrderHot(ta *entity.FcTransfersApply, coinSet *entity.FcCoinSet) (*transfer.NyzoOrderRequest, error) {
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
	coinType := s.CoinName
	if ta.Eostoken != "" {
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
	fromAddr = fromAddrs[0]
	chainBalance, err := s.getChainBalance(coinType, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("get address(%s) chain amount error:%v", fromAddr, err)
	}
	//填充参数
	orderReq := &transfer.NyzoOrderRequest{}

	chainAmount, _ := decimal.NewFromString(chainBalance)
	// 判断冷地址是否有足够的前去出账

	var fee decimal.Decimal
	if ta.Fee != "" {
		fee, _ = decimal.NewFromString(ta.Fee)
		orderReq.Fee = fee.String()
	} else {
		fee = decimal.Zero
		orderReq.Fee = "0"
	}

	transAmount := toAmount.Add(fee)
	if chainAmount.LessThan(transAmount) {
		return nil, fmt.Errorf("冷地址（%s）出账金额不足：链上金额：%s，出账所需总金额（toAmount+fee）：%s",
			fromAddr, chainAmount.String(), transAmount.String())
	}

	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = strings.ToUpper(ta.CoinName)

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Value = toAmount.String()
	orderReq.Memo = ta.Memo

	return orderReq, nil
}

func (s *NyzoTransferService) getChainBalance(coinName, address string) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	params := make(map[string]interface{})
	params["coin_name"] = coinName
	params["address"] = address
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/getBalance", cfg.Url, s.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return "", fmt.Errorf("%s get balance error: %v", coinName, err)
	}
	gbr, err := transfer.DecodeGetBalanceResp(data)
	if err != nil {
		return "", err
	}
	balance, ok := gbr.Data.(string)
	if !ok {
		return "", fmt.Errorf("%v is not string", gbr.Data)
	}
	return balance, nil
}
