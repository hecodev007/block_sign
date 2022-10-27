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
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"math/rand"
	"xorm.io/builder"
)

type QtumTransferService struct {
	CoinName string
}

func NewQtumTransferService() service.TransferService {
	return &QtumTransferService{
		CoinName: "qtum",
	}
}

func (srv *QtumTransferService) VaildAddr(address string) error {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/qtum/validateaddress?address=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	BsvResp := transfer.DecodeBsvAddressResult(data)
	if BsvResp != nil && BsvResp.Data != nil {
		if BsvResp.Data.Isvalid {
			return nil
		}
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	return err
}

func (srv *QtumTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}

func (srv *QtumTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, err := srv.getEstimateTpl(ta, workerId)
	if err != nil {
		return err
	}
	err = srv.createServiceCreate(orderReq)
	if err != nil {
		return err
	}
	return nil
}
func (srv *QtumTransferService) createServiceCreate(orderReq *transfer.QtumOrderRequest) error {
	dd, _ := json.Marshal(orderReq)
	log.Infof("qtum 发送：%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/qtum/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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

func (srv *QtumTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.QtumOrderRequest, error) {
	var (
		changeAddress, fromAddress, toAddress, token string
		coinSet                                      *entity.FcCoinSet //db币种配置
	)
	//判断是主链币还是代币
	coinName := ta.CoinName
	//设置为合约转账
	if ta.Eoskey != "" {
		coinName = ta.Eoskey
		token = ta.Eostoken
	}
	coinSet = global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少币种设置：%s", coinName)
	}
	//查询找零地址
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("Qtum 商户=[%d],查询Qtum找零地址失败", ta.AppId)
	}
	//随机选择
	randIndex := util.RandInt64(0, int64(len(changes)))
	changeAddress = changes[randIndex]

	//查找to地址
	//查询出账地址和金额
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddress = toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	toAmount := toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart()

	if token != "" {
		//查找 from地址
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
		fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinName, toAddrs[0].ToAmount).
			And(builder.In("address", coldAddrs)), 0)
		if err != nil {
			return nil, fmt.Errorf("err:%s", err.Error())
		}
		if len(fromAddrs) == 0 {
			return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddress)
		}
		fromAddress = fromAddrs[rand.Intn(len(fromAddrs))] //随机获取一个出账地址
	}
	oa := transfer.QtumOrderAddressReq{
		Address:  toAddress,
		Amount:   toAmount,
		Quantity: toAddrAmount.Shift(int32(coinSet.Decimal)).String(),
	}
	orderReq := &transfer.QtumOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = worker

	orderReq.FromAddress = fromAddress
	orderReq.ChangeAddress = changeAddress
	orderReq.Token = token
	orderReq.OrderAddress = append(orderReq.OrderAddress, oa)
	return orderReq, nil
}
