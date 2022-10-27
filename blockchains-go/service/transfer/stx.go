package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"math/rand"
	"strings"
	"sync"
	"xorm.io/builder"

	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type StxTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewStxTransferService() service.TransferService {
	return &StxTransferService{
		CoinName: "stx",
		Lock:     &sync.Mutex{},
	}
}

func (s *StxTransferService) VaildAddr(address string) error {
	if !strings.HasPrefix(address, "S") {
		return errors.New("stx address prefix is not 'S' ")
	}
	return s.validAddress(address)
	//return errors.New("do not support valid address")
}

func (s *StxTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		//mch        *entity.FcMch
		coinSet  *entity.FcCoinSet
		orderReq *transfer.StxOrderRequest
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
	orderReq, err = s.buildOrderHot(int32(coinSet.Decimal), ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	//amount = decimal.NewFromInt(orderReq.ToAddrs[0].ToAmount)

	//createData, _ = json.Marshal(orderReq)
	//orderHot := &entity.FcOrderHot{
	//	ApplyId:      ta.Id,
	//	ApplyCoinId:  coinSet.Id,
	//	OuterOrderNo: ta.OutOrderid,
	//	OrderNo:      ta.OrderId,
	//	MchName:      mch.Platform,
	//	CoinName:     ta.CoinName,
	//	FromAddress:  orderReq.FromAddress,
	//	ToAddress:    orderReq.ToAddrs[0].ToAddr,
	//	Amount:       amount.IntPart(), //转换整型
	//	Quantity:     amount.String(),
	//	Decimal:      int64(global.CoinDecimal[ta.CoinName].Decimal),
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
	//
	//orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	//err = dao.FcOrderHotInsert(orderHot)
	//if err != nil {
	//	err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
	//	// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
	//	log.Error(err.Error())
	//	// 发送给钉钉
	//	dingding.ErrTransferDingBot.NotifyStr(err.Error())
	//}
	return txid, nil
}

func (s *StxTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	return errors.New("do not support cold transfer")
}

//=================私有方法=================

func (s *StxTransferService) walletServerCreateHot(orderReq *transfer.StxOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/blockstack/Transfer", cfg.Url), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result := transfer.DecodeStxTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result["data"].(string), nil
}
func (s *StxTransferService) buildOrderHot(dec int32, ta *entity.FcTransfersApply) (*transfer.StxOrderRequest, error) {
	var (
		fromAddr   string
		changeAddr string
		toAddr     string
		toAmount   decimal.Decimal
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
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))]

	//查询找零地址
	changes, err := dao.FcGenerateAddressFindIn([]string{fromAddr})
	if err != nil {
		return nil, err
	}
	if len(changes) != 1 {
		return nil, fmt.Errorf("商户=[%d],查询%s找零地址失败", ta.AppId, s.CoinName)
	}
	//随机选择
	//randIndex := util.RandInt64(0, int64(len(changes)))
	changeAddr = changes[0].Address

	//填充参数
	orderReq := &transfer.StxOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	//orderReq. = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = strings.ToUpper(ta.CoinName)

	orderReq.FromAddress = fromAddr
	orderReq.ChanegeAddress = changeAddr
	orderAddress := transfer.StxToAddrAmount{
		ToAddr:   toAddr,
		ToAmount: toAmount.Shift(dec).IntPart(),
	}
	orderReq.ToAddrs = append(orderReq.ToAddrs, orderAddress)
	orderReq.Memo = ta.Memo
	fee, _ := decimal.NewFromString(ta.Fee)
	orderReq.TransferFee = fee.Shift(dec).IntPart()
	return orderReq, nil
}

func (s *StxTransferService) validAddress(address string) error {
	if address == "" {
		return errors.New("address is null")
	}
	url := conf.Cfg.CoinServers[s.CoinName].Url + "/conversionAddress"
	reqData := map[string]string{
		"Address":  address,
		"CoinName": "stacks",
	}
	data, err := util.PostJson(url, reqData)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", s.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	stxResp := transfer.DecodeStxTransferResp(data)
	if stxResp == nil || stxResp["message"] == nil || stxResp["data"] == nil {
		return errors.New("stx address valid error: resp data is null")
	}
	if stxResp["message"].(string) == "生成地址成功" {
		return nil
	}
	return fmt.Errorf("stx valid address error: %v", stxResp["message"])
}
