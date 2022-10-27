package transfer

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"math/rand"
	"strings"
	"xorm.io/builder"
)

type HxTransferService struct {
	CoinName string
}

func NewHxTransferService() service.TransferService {
	return &HxTransferService{CoinName: "hx"}
}
func (srv *HxTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}
func (srv *HxTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := srv.buildOrder(ta)
	if err != nil {
		return err
	}
	err = srv.walletServerCreateCold(orderReq)
	if err != nil {
		//改变表状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		return err
	}
	return nil
}
func (srv *HxTransferService) VaildAddr(address string) error {
	if !strings.HasPrefix(address, "HX") {
		return errors.New("don`t have prefix HX")
	}
	return nil
}
func (srv *HxTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.HxOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	coinName := ta.CoinName
	//设置为合约转账
	if ta.Eoskey != "" {
		coinName = ta.Eoskey
	}
	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少币种设置：%s", coinName)
	}
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
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinName, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	//填充参数
	orderReq := &transfer.HxOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(coinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(int32(coinSet.Decimal)).IntPart()
	orderReq.Memo = ta.Memo
	//构建订单数组
	var orderAddress []transfer.HxOrderAddress
	orderFrom := transfer.HxOrderAddress{
		Dir:     0,
		Address: fromAddr,
		Amount:  0,
	}
	orderAddress = append(orderAddress, orderFrom)
	orderTo := transfer.HxOrderAddress{
		Dir:     1,
		Address: toAddr,
		Amount:  toAmount.Shift(int32(coinSet.Decimal)).IntPart(),
	}
	orderAddress = append(orderAddress, orderTo)
	orderReq.OrderAddress = orderAddress
	return orderReq, nil
}
func (srv *HxTransferService) walletServerCreateCold(orderReq *transfer.HxOrderRequest) error {
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/hx/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil
}
