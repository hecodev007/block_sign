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
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

type SteemTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewSteemTransferService() *SteemTransferService {
	return &SteemTransferService{
		CoinName: "steem",
		Lock:     &sync.Mutex{},
	}
}
func (s *SteemTransferService) VaildAddr(address string) error {
	//if len(address) > 12 {
	//	return errors.New("账户长度大于12位")
	//}
	fmt.Println(address)
	return nil
}
func (s *SteemTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	fmt.Println(ta)
	panic("implement me")
}
func (s *SteemTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	coinSet := global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err := s.buildOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	txid, err = s.walletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		return "", err
	}
	createData, _ := json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      ta.Applicant,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.FromAddress,
		ToAddress:    orderReq.ToAddress,
		//Amount:       toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:   orderReq.Quantity,
		Decimal:    int64(3),
		CreateData: string(createData),
		Status:     int(status.UnknowErrorStatus),
		CreateAt:   time.Now().Unix(),
		UpdateAt:   time.Now().Unix(),
	}

	if err != nil || txid == "" {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		//写入热钱包表，创建失败
		return "", err
	}
	orderHot.Status = int(status.BroadcastStatus)
	orderHot.TxId = txid
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		//发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}

//私有方法 构建wax订单
func (s *SteemTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.OrderRequestV1, error) {
	var (
		fromAddr string
		pubKey   string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.Find(builder.Eq{
		//"type":        address.AddressTypeUser,
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
	coinType := strings.ToLower(ta.CoinName)

	coin := global.CoinDecimal[coinType]
	if coin == nil {
		return nil, fmt.Errorf("读取 %s coinSet 设置异常", coinType)
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs[0].Address)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]
	pub := new(PubKey)
	json.Unmarshal([]byte(coldAddrs[0].Json), pub)
	pubKey = pub.Key
	toAmount2, _ := toAmount.Float64()
	//填充参数
	orderReq := &transfer.OrderRequestV1{}
	//orderReq.ApplyId = int64(ta.Id)
	//orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderId = ta.OrderId
	orderReq.MchId = fmt.Sprintf("%v", ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	//orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.SignPubKey = pubKey
	//qcoin := strings.ToUpper(coinType)
	//qcoin

	orderReq.Quantity = fmt.Sprintf("%.3f", toAmount2)

	return orderReq, nil
}

//创建交易接口参数
func (s *SteemTransferService) walletServerCreate(orderReq *transfer.OrderRequestV1) (string, error) {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/transfer", conf.Cfg.HotServers[s.CoinName].Url, s.CoinName), conf.Cfg.HotServers[s.CoinName].User, conf.Cfg.HotServers[s.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	if data == nil {
		log.Error("steem请求失败")
		return "", errors.New("请求失败")
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	//type SignHeader struct {
	//	MchId    string `json:"mch_no" `
	//	MchName  string `json:"mch_name" binding:"required"`
	//	OrderId  string `json:"order_no" `
	//	CoinName string `json:"coin_name" binding:"required"`
	//}
	type TransferReturns struct {
		//SignHeader
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"txid"` //txid
	}

	result := &TransferReturns{}
	err = json.Unmarshal(data, result)
	if err != nil {
		fmt.Println("steem transfer error", err)
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderId)
	}

	//result := transfer.DecodeWalletServerRespOrder(data)
	//if result == nil {
	//	return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderId)
	//}
	if result.Code != 0 || result.Data == "" {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，OrderId：%s，err:%s", orderReq.OrderId, string(data))
	}

	return result.Data, nil
}
