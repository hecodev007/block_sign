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
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"github.com/tendermint/tendermint/libs/bech32"
	"strings"

	"sync"
	"time"
	"xorm.io/builder"
)

const DipDecimal = 12

type DipTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewDipTransferService() service.TransferService {
	return &DipTransferService{CoinName: "dip",
		Lock: &sync.Mutex{}}
}

func (s *DipTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {

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
	coinSet = global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err := s.buildOrder(ta)

	if err != nil {
		//改变表状态 外层已经定义失败状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,cocos 构建异常:%s", ta.Id, err.Error())
		//	return err
		//}
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount, _ = decimal.NewFromString(orderReq.AmountInt64)
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
	txid, err = s.walletServerCreate(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		//写入热钱包表，创建失败
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
func (s *DipTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return fmt.Errorf("don't implement for dip cold ")
}
func (s *DipTransferService) VaildAddr(address string) error {
	if len(strings.TrimSpace(address)) == 0 {
		return fmt.Errorf("address is null :%s", address)
	}
	bz, err := s.getFromBech32(address, "dip")
	if err != nil {
		return fmt.Errorf("dip address error: %v", err)
	}
	n := len(bz)

	if n == 10 || n == 20 {
		return nil
	}
	return fmt.Errorf("invalid address length %d", n)
}

//创建交易接口参数
func (s *DipTransferService) walletServerCreate(orderReq *transfer.DipOrderRequest) (txid string, err error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	log.Info(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if err != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,error:%s", orderReq.OuterOrderNo, err.Error())
	}
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	txid = result.Data.(string)
	return txid, nil
}

//======================私有方法==================
//私有方法 构建dip订单
func (s *DipTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.DipOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	coinSet := global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少dip币种信息")
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
	fromAddr = fromAddrs[0] // 这里dip只有一个
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
	orderReq := &transfer.DipOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.ApplyCoinId = int64(coinSet.Id)
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.AmountInt64 = toAmount.Shift(DipDecimal).String()
	orderReq.Denom = "pdip"
	return orderReq, nil
}

// -------------------验证地址

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func (s *DipTransferService) getFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}
