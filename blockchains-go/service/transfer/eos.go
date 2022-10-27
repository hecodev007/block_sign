package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"sync"
	"xorm.io/builder"
)

type EosTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewEosTransferService() service.TransferService {
	return &EosTransferService{
		CoinName: "eos",
		Lock:     &sync.Mutex{},
	}
}
func (s *EosTransferService) VaildAddr(address string) error {
	if len(address) > 12 {
		return errors.New("账户长度大于12位")
	}
	return nil
}
func (s *EosTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	panic("implement me")

}
func (s *EosTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := s.buildOrder(ta)
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

//私有方法 构建wax订单
func (s *EosTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.OrderRequest, error) {
	var (
		fromAddr string
		pubKey   string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.Find(builder.Eq{
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
	coinType := strings.ToLower(ta.CoinName)
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
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
	orderReq := &transfer.OrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.SignPubKey = pubKey
	qcoin := strings.ToUpper(coinType)
	if qcoin == "GBT-EOS" {
		//兼容处理，暂时硬编码
		qcoin = "GBT"
	} else if qcoin == "USDT-EOS" {
		//兼容处理，暂时硬编码
		qcoin = "USDT"
	} else if qcoin == "OGX-OLD" {
		//兼容处理，暂时硬编码
		qcoin = "OGX"
	} else if qcoin == "KEY-EOS" {
		//兼容处理，暂时硬编码
		qcoin = "KEY"
	} else if qcoin == "HOO-EOS" {
		//兼容处理，暂时硬编码
		qcoin = "HOO"
	} else if qcoin == "ADD-EOS" {
		//兼容处理，暂时硬编码
		qcoin = "ADD"
	}

	orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(coin.Decimal)+"f"+" %s", toAmount2, qcoin)
	if ta.Eostoken != "" {
		if coin.Token != ta.Eostoken {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.Token = ta.Eostoken
	} else {
		orderReq.Token = "eosio.token"
	}
	return orderReq, nil
}

//创建交易接口参数
func (s *EosTransferService) walletServerCreate(orderReq *transfer.OrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/transfer", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	if data == nil {
		log.Error("请求失败")
		return errors.New("请求失败")
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
