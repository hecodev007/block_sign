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
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"time"
	"xorm.io/builder"
)

/*
eth的 一些钉钉服务写在这里
*/

var (
	coinName    = "eth"
	coinDecimal = 18
)

// 打手续费

func EthTransferFee(mchId int64, to, mchName, feeFloat string) error {
	//0. 判断to地址是否为该
	toList, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{"address": to, "coin_name": coinName}.
		And(builder.In("type", []int{1, 2})))
	if err != nil {
		return fmt.Errorf("find to address error, %v", err)
	}
	if len(toList) != 1 {
		return fmt.Errorf("该指定商户[%d]下没有查找到币种[%s]to地址[%s]", mchId, coinName, to)
	}

	toAddr, err := dao.FcGenerateAddressGetByAddressAndMchId(to, int(mchId))
	if err != nil {
		return fmt.Errorf("find to address error, %v", err)
	}
	if toAddr == nil && (toAddr.Type != 1 && toAddr.Type != 2) {
		return fmt.Errorf("该指定商户[%d]下没有查找到该to地址[%s]", mchId, to)
	}

	var fee decimal.Decimal
	// 1. 计算需要打手续费的金额
	if feeFloat != "" {
		fee, _ = decimal.NewFromString(feeFloat)
		if fee.GreaterThan(decimal.NewFromInt(1)) {
			return fmt.Errorf("手续费大于1，fee=%s", feeFloat)
		}
		if fee.LessThan(decimal.NewFromFloat32(0.001)) {
			return fmt.Errorf("手续费小于0.001，fee=%s", feeFloat)
		}
	} else {
		fee, _ = decimal.NewFromString("0.01")
	}

	if toAddr.Type == 2 && fee.GreaterThan(decimal.NewFromFloat32(0.1)) {
		return fmt.Errorf("用户手续费不能大于0.1，目前fee=%s", feeFloat)
	}

	// 2. 根据mchId查找手续费地址
	//查找手续费地址
	feeAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"type": 3, "coin_type": "eth", "app_id": mchId}.
		And(builder.Expr("amount >= 0.003 and forzen_amount = 0")), 10)
	if err != nil {
		return err
	}
	if len(feeAddrs) == 0 {
		return errors.New("没有查找到手续费地址大于0.01eth的地址！！！")
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
		CoinName:   "eth",
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("FEE_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mchName,
		Operator:   "Robot",
		AppId:      int(mchId),
		Type:       "fee",
		Purpose:    "自动归集",
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
	if err == nil {
		//开始请求钱包服务归集
		orderReq := &transfer.EthTransferFeeReq{}
		orderReq.ApplyId = appId
		orderReq.OuterOrderNo = feeApply.OutOrderid
		orderReq.OrderNo = feeApply.OrderId
		orderReq.MchId = int64(mchId)
		orderReq.MchName = mchName
		orderReq.CoinName = "eth"
		orderReq.FromAddr = feeAddress
		orderReq.ToAddrs = []string{to}
		orderReq.NeedFee = fee.Shift(18).String() //eth -> wei
		orderReq.Worker = service.GetWorker(coinName)
		if err = walletServerFee(orderReq); err != nil {
			return fmt.Errorf("[%s] 地址大手续费错误，Err=[%v]", to, err)
		}
	} else {
		return fmt.Errorf("create app id error, %v", err)
	}
	return nil
}

func EthCollectToken(name string, mch *entity.FcMch, fromAddresses []string) error {
	//1. 查找冷地址
	toAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mch.Id,
		"coin_name":   coinName,
		"out_orderid": "collect",
	})
	if err != nil {
		return fmt.Errorf("%s find cold address error,%v", name, err)
	}
	if len(toAddrs) == 0 {
		return fmt.Errorf("%s do not find any cold address", name)
	}
	to := toAddrs[0]
	//2. 查找coin的配置
	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1, "name": name})
	if err != nil {
		return fmt.Errorf("%s find coin set error,%v", name, err)
	}
	if len(coins) != 1 {
		return fmt.Errorf("%s do not find coin set", name)
	}
	coin := coins[0]
	//3. 构建订单
	cltApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   coinName,
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mch.Platform,
		Operator:   "Robot",
		AppId:      mch.Id,
		Type:       "gj",
		Purpose:    fmt.Sprintf("%s自动归集", coin.Name),
		Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}
	if name != coinName {
		cltApply.Eostoken = coin.Token
		cltApply.Eoskey = coin.Name
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     to,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	collectAddrs := make([]string, 0)
	for _, from := range fromAddresses {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from,
			AddressFlag: "from",
			Status:      0,
			Lastmodify:  cltApply.Lastmodify,
		})
		collectAddrs = append(collectAddrs, from)

		if strings.ToLower(from) == to {
			return fmt.Errorf("%s from和to为同一地址", from)
		}
	}

	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		return fmt.Errorf("build app id error,%v", err)
	}
	//开始请求钱包服务归集
	orderReq := &transfer.EthCollectReq{}
	orderReq.ApplyId = appId
	orderReq.OuterOrderNo = cltApply.OutOrderid
	orderReq.OrderNo = cltApply.OrderId
	orderReq.MchId = int64(mch.Id)
	orderReq.MchName = mch.Platform
	orderReq.CoinName = coinName
	orderReq.FromAddrs = collectAddrs
	orderReq.ToAddr = to
	if name != coinName { //如果是代币归集
		orderReq.ContractAddr = coin.Token
		orderReq.Decimal = coin.Decimal
	}
	if err := walletServerCollect(orderReq); err != nil {
		return fmt.Errorf("%s 归集失败，Err： %v", name, err)
	}
	return nil
}

func EthInternal(mch *entity.FcMch, amount string, fromAddress, toAddress string) error {
	var (
		toAddr   *entity.FcGenerateAddressList
		fromAddr *entity.FcGenerateAddressList
	)

	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}

	addrList, err := dao.FcGenerateAddressFindInternal(mch.Id, coinName, []string{fromAddress, toAddress})
	if err != nil {
		return err
	}
	for _, a := range addrList {
		if a.Address == fromAddress {
			fromAddr = a
		}
		if a.Address == toAddress {
			toAddr = a
		}
	}
	if toAddr == nil {
		return fmt.Errorf("to%s 地址不合法", toAddress)
	}
	if fromAddr == nil {
		return fmt.Errorf("from%s 地址不合法", fromAddress)
	}
	to := toAddr.Address
	from := fromAddr.Address

	//3. 构建订单
	cltApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   coinName,
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mch.Platform,
		Operator:   "Robot",
		AppId:      mch.Id,
		Type:       "gj",
		Purpose:    fmt.Sprintf("%s内部主链币转账", coinName),
		Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
	}
	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     to,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})

	collectAddrs := make([]string, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     from,
		AddressFlag: "from",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	collectAddrs = append(collectAddrs, from)

	if strings.ToLower(from) == to {
		return fmt.Errorf("%s from和to为同一地址", from)
	}

	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		return fmt.Errorf("build app id error,%v", err)
	}
	//开始请求钱包服务归集
	orderReq := &transfer.EthCollectReq{}
	orderReq.ApplyId = appId
	orderReq.OuterOrderNo = cltApply.OutOrderid
	orderReq.OrderNo = cltApply.OrderId
	orderReq.MchId = int64(mch.Id)
	orderReq.MchName = mch.Platform
	orderReq.CoinName = coinName
	orderReq.FromAddrs = collectAddrs
	orderReq.ToAddr = to
	orderReq.Amount = amountDecimal.Shift(int32(coinDecimal)).String()
	if err := walletServerCollect(orderReq); err != nil {
		return fmt.Errorf("%s 内部转账，Err： %v", coinName, err)
	}
	return nil
}

func walletServerFee(orderReq *transfer.EthTransferFeeReq) error {
	cfg := conf.Cfg.Walletserver
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/fee", cfg.Url, coinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s fee send :%s", coinName, string(dd))
	log.Infof("%s fee resp :%s", coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerFee 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerFee 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//创建交易接口参数
func walletServerCollect(orderReq *transfer.EthCollectReq) error {
	cfg := conf.Cfg.Walletserver
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/collect", cfg.Url, coinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", coinName, string(dd))
	log.Infof("%s Collect resp :%s", coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}
