package dingding

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"strings"
	"time"
	"xorm.io/builder"
)

var transferFeeCoins = []string{"eth", "heco", "bsc"}

func InDingSupportArray(coinName string) bool {
	for _, name := range transferFeeCoins {
		if strings.ToLower(coinName) == strings.ToLower(name) {
			return true
		}
	}
	return false
}
func CoinTransferFee(mchId int64, coinName, to, mchName, feeFloat string) error {
	//1. 判断to地址是否为该冷地址或者用户地址
	toList, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{"address": to, "coin_name": coinName}.
		And(builder.In("type", []int{1, 2})))
	if err != nil {
		return fmt.Errorf("find to address error, %v", err)
	}
	if len(toList) != 1 {
		return fmt.Errorf("该指定商户[%d]下没有查找到该to地址[%s]", mchId, to)
	}

	fee, _ := decimal.NewFromString(feeFloat)
	minFeeAmount := fee.Mul(decimal.NewFromInt(2)) //手续费地址最小是要打手续费的两倍，不然不让打手续费
	mfa, _ := minFeeAmount.Float64()
	// 2. 根据mchId查找手续费地址
	//查找手续费地址
	feeAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 3, "coin_type": coinName, "app_id": mchId}.
		And(builder.Expr("amount >= ? and forzen_amount = 0", mfa)), 10)
	if err != nil {
		return err
	}
	if len(feeAddrs) == 0 {
		return fmt.Errorf("没有查找到手续费地址大于%s%s的地址！！！", minFeeAmount.String(), coinName)
	}
	//3. 取一个最佳的手续费地址
	var feeAddress = ""
	for _, f := range feeAddrs {
		amount, _ := decimal.NewFromString(f.Amount)
		pendingAmount, _ := decimal.NewFromString(f.PendingAmount)
		useAmount := amount.Sub(pendingAmount)
		if useAmount.GreaterThan(fee) {
			feeAddress = f.Address
			break
		}
	}

	//生成订单
	feeApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   coinName,
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("FEE_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mchName,
		Operator:   "Robot",
		AppId:      int(mchId),
		Type:       "fee",
		Purpose:    "打手续费",
		Status:     int(entity.ApplyStatus_Fee), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     feeAddress,
		AddressFlag: "from",
		Status:      0,
	})
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     to,
		AddressFlag: "to",
		Status:      0,
	})
	appId, err := feeApply.TransactionAdd(applyAddresses)

	//4. 根据每一个币种去做对应的打手续费
	srv := GetIDingService(strings.ToLower(coinName))
	return srv.TransferFee(feeAddress, to, appId, feeApply, fee)
}

func CoinCollectToken(coinName, coldAddress string, mch *entity.FcMch, fromAddresses []string) error {

	//1. 查找coin的配置
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1, "name": coinName})
	if err != nil {
		return fmt.Errorf("%s find coin set error,%v", coinName, err)
	}
	if len(coins) != 1 {
		return fmt.Errorf("%s do not find coin set", coinName)
	}
	coin := coins[0]
	//2. 查找主链
	mainCoin, err := dao.FcCoinSetGetByStatus(coin.Pid, 1)
	if err != nil {
		return fmt.Errorf("没有找到该币种%s的主链信息：%v", coinName, err)
	}
	//  要判断主链是否支持该功能，不然的话，后面反射会panic
	if !InDingSupportArray(strings.ToLower(mainCoin.Name)) {
		return fmt.Errorf("不支持该币种%s的代币归集服务。", mainCoin.Name)
	}
	//根据主链的名字去查找冷地址
	var to string
	mainName := mainCoin.Name
	if coldAddress != "" {
		// 判断是否是冷地址
		//1. 判断to地址是否为该冷地址或者用户地址
		toList, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{"address": coldAddress, "coin_name": mainName}.
			And(builder.In("type", []int{1})))
		if err != nil {
			return fmt.Errorf("%s查找指定的冷地址失败：%v", mainName, err)
		}
		if len(toList) != 1 {
			return fmt.Errorf("%s未找到指定的冷地址：%s", mainName, coldAddress)
		}

		to = coldAddress
	} else {
		// 查找冷地址
		toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
			"type":        address.AddressTypeCold,
			"status":      address.AddressStatusAlloc,
			"platform_id": mch.Id,
			"coin_name":   mainName,
		})
		if err != nil {
			return fmt.Errorf("%s find cold address error,%v", mainName, err)
		}
		if len(toAddrs) == 0 {
			return fmt.Errorf("%s do not find any cold address", mainName)
		}
		to = toAddrs[0]
	}

	srv := GetIDingService(mainName)
	return srv.CollectToken(mainName, to, mch, fromAddresses, coin)
}

func CoinFindAddressFee(mainName, address string, mch *entity.FcMch) (dbAmount, chainAmount string, err error) {
	//1. 判断地址是否为该冷地址或者用户地址
	toList, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{"address": address, "coin_name": mainName, "platform_id": mch.Id}.
		And(builder.In("type", []int{1, 2})))
	if err != nil {
		return "", "", fmt.Errorf("%s find to address error, %v", mainName, err)
	}
	if len(toList) != 1 {
		return "", "", fmt.Errorf("该指定商户[%d]下没有查找到该to地址[%s]", mch.Id, address)
	}

	//2. 获取数据库金额
	addresses, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"address": address, "coin_type": mainName, "app_id": mch.Id}, 10)
	if err != nil {
		return "", "", fmt.Errorf("币种%s查询地址%s余额错误：%v", mainName, address, err)
	}
	if len(addresses) != 1 {
		return "", "", fmt.Errorf("币种%s未查找到地址%s余额", mainName, address)
	}
	dbAmount = addresses[0].Amount
	srv := GetIDingService(mainName)
	chainAmount, err = srv.FindCoinFee(mainName, address, mch)
	if err != nil {
		return "", "", fmt.Errorf("获取链上余额错误： %v", err)
	}
	return
}
