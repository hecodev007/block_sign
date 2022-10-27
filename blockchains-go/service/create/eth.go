package create

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/proto"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type EthCreator struct {
	serverUrl string
}

func NewEthCreator(url string) *EthCreator {
	return &EthCreator{
		serverUrl: url,
	}
}

func (s *EthCreator) Name() string {
	return "eth"
}

func (s *EthCreator) CreateTx(request *entity.FcTransfersApply) (int64, error) {
	var (
		res *proto.OrderRequest
		err error
	)

	isToken := request.Eostoken != "" || request.Eoskey != ""
	//判断订单的状态
	switch request.Type {
	case "cz", "fee":
		//判断币种精度
		if isToken {
			res, err = s.transfer(request, request.Eoskey, request.Eostoken)
		} else {
			res, err = s.transfer(request, request.CoinName, "")
		}
	case "gj":
		if isToken {
			res, err = s.collect(request, request.Eoskey, request.Eostoken)
		} else {
			res, err = s.transfer(request, request.CoinName, "")
		}
	default:
		return -1, fmt.Errorf("don't support tx type %s", request.Type)
	}

	if err != nil {
		return -1, err
	}

	return service.HttpCreateTx(s.serverUrl, res)
}

func (s *EthCreator) transfer(request *entity.FcTransfersApply, coinName, contractAddress string) (*proto.OrderRequest, error) {

	coinSet, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("不支持的代币 %s", coinName))
		return nil, err
	}

	toAddress, err := dao.GetApplyAddressByApplyCoinId(int64(request.Id), "to")
	if err != nil {
		return nil, err
	}

	if toAddress == nil || toAddress.Address == "" {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("接收地址不能为空"))
		return nil, err
	}

	amount, err := decimal.NewFromString(toAddress.ToAmount)
	if err != nil {
		return nil, err
	}

	fromAddress, err := dao.FcAddressAmountGetCloudAddress(1, 2, request.AppId, coinName, toAddress.ToAmount)
	if err != nil {
		return nil, err
	}

	if fromAddress == nil || fromAddress.Address == "" {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("%s .. %s 出账地址余额不足", coinName, coinName))
		return nil, fmt.Errorf("%s .. %s 出账地址余额不足", coinName, coinName)
	}

	if toAddress.Address == fromAddress.Address {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("发送地址和接收地址不能相同 "))
		return nil, fmt.Errorf("发送地址和接收地址不能相同")
	}

	if !amount.IsPositive() {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("出账金额不合法 %v", amount))
		return nil, fmt.Errorf("出账金额不合法 %v", amount)
	}

	totalAmount, err := dao.FcAddressAmountGetTotalAmount(request.AppId, coinName)
	if err != nil {
		return nil, err
	}

	if !totalAmount.IsPositive() {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("ETH代币 %s 商户余额不足", coinName))
		return nil, fmt.Errorf("ETH代币 %s 商户余额不足", coinName)
	}

	if amount.Cmp(totalAmount) < 0 {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("商户ETH代币 %s 余额不足,不够出账", coinName))
		return nil, fmt.Errorf("商户ETH代币 %s 余额不足,不够出账", coinName)
	}

	param := &proto.OrderRequest{
		ApplyId:      int64(request.Id),
		ApplyCoinId:  int64(coinSet.Id),
		OuterOrderNo: request.OutOrderid,
		MchName:      request.Applicant,
		CoinName:     coinName,
		Token:        contractAddress,
		Decimal:      int32(coinSet.Decimal),
		Worker:       service.GetWorker(request.CoinName),
		OrderAddress: make([]*proto.OrderAddrRequest, 0),
	}

	param.OrderAddress = append(param.OrderAddress, &proto.OrderAddrRequest{
		Dir:         1,
		Address:     toAddress.Address,
		TokenAmount: toAddress.ToAmount,
	})

	param.OrderAddress = append(param.OrderAddress, &proto.OrderAddrRequest{
		Dir:     0,
		Address: fromAddress.Address,
	})

	return param, nil
}

func (s *EthCreator) collect(request *entity.FcTransfersApply, coinName, contractAddress string) (*proto.OrderRequest, error) {

	coinSet, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("不支持的代币 %s", coinName))
		return nil, err
	}

	toAddress, err := dao.GetApplyAddressByApplyCoinId(int64(request.Id), "to")
	if err != nil {
		return nil, err
	}

	if toAddress == nil || toAddress.Address == "" {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("接收地址不能为空"))
		return nil, err
	}

	amount, err := decimal.NewFromString(toAddress.ToAmount)
	if err != nil {
		return nil, err
	}

	fromAddress, err := dao.FcAddressAmountGetCloudAddress(1, 2, request.AppId, coinName, toAddress.ToAmount)
	if err != nil {
		return nil, err
	}

	if fromAddress == nil || fromAddress.Address == "" {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("%s .. %s 出账地址余额不足", coinName, coinName))
		return nil, fmt.Errorf("%s .. %s 出账地址余额不足", coinName, coinName)
	}

	if toAddress.Address == fromAddress.Address {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("发送地址和接收地址不能相同 "))
		return nil, fmt.Errorf("发送地址和接收地址不能相同")
	}

	if !amount.IsPositive() {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("出账金额不合法 %v", amount))
		return nil, fmt.Errorf("出账金额不合法 %v", amount)
	}

	totalAmount, err := dao.FcAddressAmountGetTotalAmount(request.AppId, coinName)
	if err != nil {
		return nil, err
	}

	if !totalAmount.IsPositive() {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("ETH代币 %s 商户余额不足", coinName))
		return nil, fmt.Errorf("ETH代币 %s 商户余额不足", coinName)
	}

	if amount.Cmp(totalAmount) < 0 {
		service.NotifyCode(request.Applicant, coinName, request.OutOrderid, fmt.Sprintf("商户ETH代币 %s 余额不足,不够出账", coinName))
		return nil, fmt.Errorf("商户ETH代币 %s 余额不足,不够出账", coinName)
	}

	param := &proto.OrderRequest{
		ApplyId:      int64(request.Id),
		ApplyCoinId:  int64(coinSet.Id),
		OuterOrderNo: request.OutOrderid,
		MchName:      request.Applicant,
		CoinName:     coinName,
		Token:        contractAddress,
		Decimal:      int32(coinSet.Decimal),
		Worker:       service.GetWorker(request.CoinName),
		OrderAddress: make([]*proto.OrderAddrRequest, 0),
	}

	param.OrderAddress = append(param.OrderAddress, &proto.OrderAddrRequest{
		Dir:         1,
		Address:     toAddress.Address,
		TokenAmount: "-8",
	})

	param.OrderAddress = append(param.OrderAddress, &proto.OrderAddrRequest{
		Dir:     0,
		Address: fromAddress.Address,
	})
	return param, nil
}
