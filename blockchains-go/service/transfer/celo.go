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

const CeloDecimal = 18

type CeloTransferService struct {
	CoinName string
	Lock     *sync.Mutex
	//nonceCtl   map[string]int		// address-->nonce
	lastFromAddress string //上一笔交易使用的地址
}

func NewCeloTransferService() service.TransferService {
	return &CeloTransferService{
		CoinName:        "celo",
		Lock:            &sync.Mutex{},
		lastFromAddress: "",
	}
}

func (s *CeloTransferService) VaildAddr(address string) error {
	if !common.IsHexAddress(address) {
		return errors.New("valid Celo address error")
	}
	return nil
}

func (s *CeloTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.CeloOrderRequest
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

func (s *CeloTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support cold transfer")
}

func (s *CeloTransferService) walletServerCreateHot(orderReq *transfer.CeloOrderRequest) (string, error) {
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
	result := transfer.DecodeCeloTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result["data"].(string), nil
}
func (s *CeloTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.CeloOrderRequest, error) {
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
	//fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	// change by jun 2020/5/20
	fromAddr = s.getFromAddress(s.lastFromAddress, fromAddrs)
	//填充参数
	orderReq := &transfer.CeloOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	// orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(int32(CeloDecimal)).String()
	// orderReq.GasLimit = uint64(60000)		// 小币种限制死没多大关系
	return orderReq, nil
}

/*
去除掉上一笔使用过的地址，使用一个新的地址出账
*/
func (s *CeloTransferService) getFromAddress(lastFromAddress string, addr []string) string {
	var fromAddress string
	if len(addr) == 1 {
		fromAddress = addr[0]
		lastFromAddress = fromAddress
		return fromAddress
	}
	var tmpAdds []string
	for _, address := range addr {
		if strings.Compare(address, lastFromAddress) == 0 {
			continue
		}
		tmpAdds = append(tmpAdds, address)
	}
	fromAddress = tmpAdds[rand.Intn(len(tmpAdds))]
	lastFromAddress = fromAddress
	return fromAddress
}
