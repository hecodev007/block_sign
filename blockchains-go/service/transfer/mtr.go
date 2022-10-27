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
	"strconv"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"

	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

const MtrDecimal = 18

type MtrTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewMtrTransferService() service.TransferService {
	return &MtrTransferService{
		CoinName: "mtr",
		Lock:     &sync.Mutex{},
	}
}

func (s *MtrTransferService) VaildAddr(address string) error {
	isOk := common.IsHexAddress(address)
	if !isOk {
		return errors.New("valid address error")
	}
	return nil
}

func (s *MtrTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		orderReq    *transfer.MtrOrderRequest
		amountInt64 decimal.Decimal //发送金额
		createData  []byte          //构造交易信息
		mch         *entity.FcMch
		coinSet     *entity.FcCoinSet
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

	orderReq, err = s.buildOrderHot(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	////写死了
	amountInt64, _ = decimal.NewFromString(orderReq.ToAmountInt64)
	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.FromAddr,
		ToAddress:    orderReq.ToAddr,
		Amount:       amountInt64.IntPart(), //转换整型
		Quantity:     amountInt64.Shift(int32(coinSet.Decimal)).String(),
		Decimal:      int64(coinSet.Decimal),
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

func (s *MtrTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support cold transfer")
}

//=================私有方法=================

func (s *MtrTransferService) walletServerCreateHot(orderReq *transfer.MtrOrderRequest) (string, error) {
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
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.New("json unmarshal response data error")
	}
	if resp["data"] == nil {
		return "", fmt.Errorf("mtr transfer error,Err=%s", string(data))
	}
	return resp["data"].(string), nil
}
func (s *MtrTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.MtrOrderRequest, error) {
	var (
		fromAddr      string
		toAddr        string
		toAmount      decimal.Decimal
		toAmountInt64 decimal.Decimal
		tokenId       int64
		//coinSet  *entity.FcCoinSet
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
	toAddrs, err1 := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err1 != nil {
		return nil, err1
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

	//如果 有合约地址，就使用合约地址
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	} else {
		coinType = s.CoinName
	}
	//coinSet = global.CoinDecimal[coinType]
	//if coinSet == nil {
	//	return nil, fmt.Errorf("缺少币种信息")
	//}

	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0",
		coinType, toAddrs[0].ToAmount).And(builder.In("address", coldAddrs)), 10)

	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址

	if ta.Eostoken != "" { //如果是代币转账
		coin := global.CoinDecimal[ta.Eoskey]
		if coin == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", ta.Eoskey)
		}
		if coin.Token != ta.Eostoken {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		//代币
		toAmountInt64 = toAmount.Shift(int32(coin.Decimal))
		tokenInt64, err := strconv.ParseInt(coin.Token, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("代币ID转换异常 %s ", coin.Token)
		}
		tokenId = tokenInt64
	} else {
		//mtr
		tokenId = 0
		toAmountInt64 = toAmount.Shift(18)
	}

	//填充参数
	orderReq := &transfer.MtrOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			CoinName: s.CoinName,
		},
		FromAddr:      fromAddr,
		ToAddr:        toAddr,
		ToAmountInt64: toAmountInt64.String(),
		FeeAddr:       "",
		Token:         tokenId,
	}
	return orderReq, nil
}
