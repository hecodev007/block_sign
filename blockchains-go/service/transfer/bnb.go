package transfer

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"xorm.io/builder"
)

type BnbTransferService struct {
	CoinName string
}

func NewBnbTransferService() service.TransferService {
	return &BnbTransferService{CoinName: "bnb"}
}

func (srv *BnbTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}
func (srv *BnbTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, err := srv.buildOrder(ta, workerId)
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
func (srv *BnbTransferService) VaildAddr(address string) error {
	//url:=conf.Cfg.CoinServers[srv.CoinName].User+"/api/v1/bnb/validateaddress?address=%s"
	//url = fmt.Sprintf(url, address)
	//data, err := util.Get(url)
	//if err != nil {
	//	err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
	//	return err
	//}
	//log.Infof("验证地址返回结果：%s", string(data))
	//bnbResp,err:=transfer.DecodeBNBAddressResp(data)
	//if err != nil {
	//	log.Errorf("验证地址错误，%s,address:%s,err=[%s]", srv.CoinName, address, err.Error())
	//	return fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	//}
	//if bnbResp.Data == "true" {
	//	return nil
	//} else {
	//	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	//	return err
	//}

	if len(address) == 0 {
		return errors.New("valid address error, must provide an address")
	}
	hrp, _, err := decodeAndConvert(address)
	if err != nil {
		return err
	}
	if hrp != "bnb" {
		return fmt.Errorf("invalid bech32 prefix,Excepted bnb,Got %s", hrp)
	}
	return nil
}

//创建交易接口参数
func (srv *BnbTransferService) walletServerCreateCold(orderReq *transfer.BNBOrderRequest) error {
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/bnb/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order 请求下单接口失败，send data:%s，outOrderId：%s，from: %s,to； %s: amount: %s",
			string(data), orderReq.OuterOrderNo, orderReq.FromAddress, orderReq.ToAddress, orderReq.Quantity)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s：，from: %s,to； %s: amount: %s",
			string(data), orderReq.OuterOrderNo, orderReq.FromAddress, orderReq.ToAddress, orderReq.Quantity)
	}
	return nil
}
func (srv *BnbTransferService) buildOrder(ta *entity.FcTransfersApply, worker string) (*transfer.BNBOrderRequest, error) {
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
		return nil, fmt.Errorf("bnb 缺少币种设置：%s", coinName)
	}
	//查找from地址与金额
	// 查找from地址和金额
	//coldAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Eq{
	//	"type":        address.AddressTypeCold,
	//	"status":      address.AddressStatusAlloc,
	//	"platform_id": ta.AppId,
	//	"coin_name":   ta.CoinName,
	//},10)
	coldAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and type = ? and app_id = ? and amount > 0.000375",
		ta.CoinName, address.AddressTypeCold, ta.AppId), 10)
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

	//fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinName, toAddrs[0].ToAmount).
	//	And(builder.In("address", coldAddrs)), 0)

	fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Expr("coin_type = ? ", coinName).
		And(builder.In("address", coldAddrs)), 10)
	if err != nil {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,find from address err:%s,", ta.Id, ta.OutOrderid, err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("实际执行币种：%s\noutorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", coinName, ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	var (
		idx         = -1
		totalAmount decimal.Decimal
	)
	for i, f := range fromAddrs {
		fromAmount, err := decimal.NewFromString(f.Amount)
		if err != nil {
			continue
		}
		totalAmount = totalAmount.Add(fromAmount)
		if fromAmount.GreaterThanOrEqual(toAmount) {
			idx = i
			break
		}
	}
	if totalAmount.IsZero() {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找from地址总金额错误,总金额：0", ta.Id, ta.OutOrderid)
	}
	if idx < 0 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,没有满足出账金额的from地址,from地址总金额：%s,出账金额：%s", ta.Id, ta.OutOrderid, totalAmount.String(), toAmount.String())
	}
	//设置出账地址
	fromAddr = fromAddrs[idx].Address
	//填充参数
	orderReq := new(transfer.BNBOrderRequest)
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = worker

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.Token = strings.ToUpper(coinSet.Name)
	//历史遗留问题，需要特殊更改
	//if strings.ToLower(coinName) == "bnb" {
	//	orderReq.Token = "BNB"
	//} else {
	//	orderReq.Token = strings.ToUpper(coinSet.Name)
	//}
	orderReq.Quantity = toAmount.Shift(int32(coinSet.Decimal)).String()
	return orderReq, nil
}

func decodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.Decode(bech)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32 error,Err=%v", err)
	}
	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("convert bits bech32 error,Err=%v", err)
	}
	return hrp, converted, nil
}
