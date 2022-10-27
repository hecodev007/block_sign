package transfer

import (
	"encoding/json"
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
	"math/rand"
	"strings"
	"sync"
	"xorm.io/builder"
)

type SeekTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewSeekTransferService() service.TransferService {
	return &SeekTransferService{
		CoinName: "seek",
		Lock:     &sync.Mutex{},
	}
}

func (s *SeekTransferService) VaildAddr(address string) error {

	data, err := util.Get(fmt.Sprintf("%s/%s/validaddress?address=%s", conf.Cfg.Walletserver.Url, s.CoinName, address))
	if err != nil {
		return err
	}
	log.Infof("vaild address :%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("请求验证接口失败，address：%s", address)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return fmt.Errorf("请求验证接口失败,服务器返回异常，address：%s", address)
	}
	res, ok := result.Data.(bool)
	if !ok {
		return fmt.Errorf("data type err")
	}
	if !res {
		return fmt.Errorf("address is avalid")
	}
	return nil
}

func (s *SeekTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	panic("implement me")
}

func (s *SeekTransferService) TransferCold(ta *entity.FcTransfersApply) error {
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

//=================私有方法=================
//创建交易接口参数
func (s *SeekTransferService) walletServerCreate(orderReq *transfer.OrderRequest2) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
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

//私有方法 构建eth订单
func (s *SeekTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.OrderRequest2, error) {
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
	orderReq := &transfer.OrderRequest2{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.OrderAddress = append(orderReq.OrderAddress, &transfer.OrderAddrRequest2{
		Dir:     transfer.DirTypeFrom,
		Address: fromAddr,
	})
	orderReq.OrderAddress = append(orderReq.OrderAddress, &transfer.OrderAddrRequest2{
		Dir:     transfer.DirTypeTo,
		Address: toAddr,
		Amount:  toAmount,
	})
	if ta.Eostoken != "" { //如果是代币转账
		coin := global.CoinDecimal[ta.Eoskey]
		if coin == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", ta.Eoskey)
		}
		if strings.ToLower(coin.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.Token = ta.Eostoken
	}
	return orderReq, nil
}
