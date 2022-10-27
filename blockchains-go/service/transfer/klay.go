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
	"sync"
	"xorm.io/builder"
)

const KlayDecimal = 18

type KlayTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewKlayTransferService() service.TransferService {
	return &KlayTransferService{
		CoinName: "klay",
		Lock:     &sync.Mutex{},
	}
}
func (s *KlayTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
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
	txid, err = s.walletServerCreate(orderReq)
	if err != nil {
		//改变表状态 外层已经统一处理
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,cocos 获取发送交易异常:%s", ta.Id, err.Error())
		//	return err
		//}
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		//写入热钱包表，创建失败
		return "", err
	}
	return txid, nil
}
func (s *KlayTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return fmt.Errorf("don't implement for kava cold ")
}
func (s *KlayTransferService) VaildAddr(address string) error {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.Get(fmt.Sprintf("%s/%s/validaddress?address=%s", cfg.Url, s.CoinName, address))
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

//创建交易接口参数
func (s *KlayTransferService) walletServerCreate(orderReq *transfer.KlayOrderRequest) (txid string, err error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}

//======================私有方法==================
//私有方法 构建kava订单
func (s *KlayTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.KlayOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	coinSet := global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少klay币种信息")
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
	fromAddr = fromAddrs[0] // 这里cocos只有一个
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
	orderReq := &transfer.KlayOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.ApplyCoinId = int64(coinSet.Id)
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Data = transfer.KlayPaymentRequest{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      toAmount.Shift(KlayDecimal).String(), //转换成
	}
	return orderReq, nil
}
