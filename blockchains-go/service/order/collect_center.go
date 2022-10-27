package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

const (
	collectTaskSuccessCode = 0
)

type CollectTaskRequest struct {
	OuterOrderNo string   `json:"outerOrderNo"`
	Address      []string `json:"address"`
	Contract     string   `json:"contract"`
	CoinCode     string   `json:"coinCode"`
	NeedAmount   string   `json:"needAmount"`
}

type CollectTaskResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewCollectEnable() bool {
	return true
}

func IsNewCollectVersion(chain string) bool {
	return chain == "eth" || chain == "trx" || chain == "bsc" || chain == "heco" || chain == "hsc"
}

func ManualCollect(mchId int, outerOrderNo string, coinType, contract, collectThreshold string, amount decimal.Decimal) (string, error) {
	pickCount := 100
	pickedAddrs, enoughAmt, _, err := getLessThanEnoughAmountList(mchId, coinType, collectThreshold, amount, pickCount)
	if err != nil {
		return "", fmt.Errorf("getLessThanEnoughAmountList 出错: %v", err)
	}
	log.Infof("ManualCollect %s 准备进行归集 %d 个地址的金额", outerOrderNo, len(pickedAddrs))
	m := ""
	if len(pickedAddrs) == 0 {
		return "", fmt.Errorf("没有可归集的地址，mchId=%d %s当前归集阈值:%s", mchId, coinType, collectThreshold)
	}

	if !enoughAmt {
		m = "\n归集后的金额可能还不足以出账，请留意链上金额，检查是否需要再次执行归集"
	}

	msg := fmt.Sprintf("订单归集请求成功\n订单:%s\n金额:%s\n选取了%d个地址来进行归集%s", outerOrderNo, amount, len(pickedAddrs), m)
	return msg, CallCollectCenter(outerOrderNo, fetchAddress(pickedAddrs), contract, coinType, amount)

}

func CheckIfNeedCollect(mchId int, outerOrderNo string, chain, coinCode, contract, collectThreshold string, amount decimal.Decimal) (needCollect bool, err error) {
	log.Infof("CheckIfNeedCollect mchId=%d outerOrderNo=%s chain=%s coinCode=%s contract=%s collectThreshold=%s amount=%s", mchId, outerOrderNo, chain, coinCode, contract, collectThreshold, amount.String())
	coinType := chain
	if coinCode != "" {
		coinType = coinCode
	}
	pickCount := 10
	pickColdAndUserAddrEnoughCount := 30

	if chain == "trx" && coinCode == "mof" {
		maxColdAddrAmt, err := getMaxAddressAmount(mchId, chain, coinType)
		if err != nil {
			return false, fmt.Errorf("获取最大余额出账地址出错: %v", err)
		}
		if maxColdAddrAmt.GreaterThanOrEqual(amount) {
			return false, nil
		}
		// 还差多少
		amount = amount.Sub(maxColdAddrAmt)
		collectThreshold = "200000"
	}

	//totalAmount := decimal.Zero
	// 判断10笔的总额是否大于出账金额
	// 从10笔内选出尽可能少的笔数进行归集
	pickedAddrs, enoughAmt, _, err := getLessThanEnoughAmountList(mchId, coinType, collectThreshold, amount, pickCount)
	if err != nil {
		return false, fmt.Errorf("getLessThanEnoughAmountList 出错: %v", err)
	}
	log.Infof("CheckIfNeedCollect %s 需要出账金额 %s，找到 %d 个需要归集的地址: %v", outerOrderNo, amount.String(), len(pickedAddrs), pickedAddrs)
	if enoughAmt {
		log.Infof("CheckIfNeedCollect %s 准备进行归集 %d 个地址的金额", outerOrderNo, len(pickedAddrs))
		return true, CallCollectCenter(outerOrderNo, fetchAddress(pickedAddrs), contract, coinType, amount)
	}
	log.Infof("CheckIfNeedCollect %s %d笔的总额小于出账金额，需要判断是否有大于出账金额的单个用户地址", outerOrderNo, pickCount)

	// 如果10笔的总额小于出账金额
	// 判断是否有大于出账金额的待归集资金
	near, err := dao.GetNearOutAmount(mchId, coinType, amount.String(), collectThreshold)
	if err != nil {
		return false, fmt.Errorf("dao.GetNearOutAmount 出错: %v", err)
	}
	if near != nil {
		log.Infof("CheckIfNeedCollect %s 找到大于出账金额的单个用户地址，准备进行归集 address(%s) amount(%s)", outerOrderNo, near.Address, near.Amount)
		return true, CallCollectCenter(outerOrderNo, []string{near.Address}, contract, coinType, amount)
	}

	log.Infof("CheckIfNeedCollect %s 没有找到大于出账金额单个用户地址，准备判断出账地址金额是否足够", outerOrderNo)
	coldAddrAmountEnough, err := checkColdAddressAmountEnough(mchId, chain, coinType, amount)
	if err != nil {
		return false, fmt.Errorf("检查出账地址余额是否足够出错: %v", err)
	}
	if coldAddrAmountEnough {
		log.Infof("CheckIfNeedCollect %s 出账地址有足够的金额出账，进入旧的正常出账流程", outerOrderNo)
		return false, nil
	}

	// 出账地址没有足够的余额
	// 从用户地址获取最大余额的30条记录，加上出账地址的余额出账
	maxColdAddrAmt, err := getMaxAddressAmount(mchId, chain, coinType)
	if err != nil {
		return false, fmt.Errorf("获取最大余额出账地址出错: %v", err)
	}
	log.Infof("CheckIfNeedCollect %s 获取到最多余额的出账地址金额 %s", outerOrderNo, maxColdAddrAmt.String())

	userAddrList, enough, err := coldAndUserAddrEnough(amount, mchId, coinType, maxColdAddrAmt, pickColdAndUserAddrEnoughCount)
	if err != nil {
		return false, fmt.Errorf("coldAndUserAddrEnough出错: %v", err)
	}
	if enough {
		log.Infof("CheckIfNeedCollect %s 出账地址有[没有]足够的金额出账，从用户地址获取最大余额的30条做归集+出账地址余额出账", outerOrderNo)
		return true, CallCollectCenter(outerOrderNo, fetchAddress(userAddrList), contract, coinType, amount)
	}

	// 计算归集了还是不足以出账，但是还是可以进行归集的
	CallCollectCenter(outerOrderNo, fetchAddress(userAddrList), contract, coinType, amount)

	msg := fmt.Sprintf("[归集报警]\n订单：%s\n出账金额：%s\n前30个用户地址总额不足出账\n出账地址余额不足出账\n请联系财务或做市补充", outerOrderNo, amount.String())
	log.Info(msg)
	dingding.ErrTransferDingBot.NotifyStr(msg)
	//go dao.ReportCollectAlert(int64(mchId), outerOrderNo, chain, coinType, amount.String())
	//
	//hungryCount := 30
	//log.Infof("CheckIfNeedCollect %s 出账地址[没有]足够的金额出账（出账金额 %s），准备获取 %d 个用户地址来归集", outerOrderNo, amount.String())
	//pickedAddrs, enoughAmt, totalAmount, err = getLessThanEnoughAmountList(mchId, coinType, collectThreshold, amount, hungryCount)
	//if err != nil {
	//	return false, err
	//}
	//if enoughAmt {
	//	log.Infof("CheckIfNeedCollect %s 准备进行归集 %d 个地址的金额", outerOrderNo, len(pickedAddrs))
	//	return true, CallCollectCenter(outerOrderNo, fetchAddress(pickedAddrs), contract, coinType, amount)
	//}
	//
	//msg = fmt.Sprintf("[归集报警]\n订单：%s\n归集%d个地址仍然不足以出账，还差%s\n请联系财务或做市补充", outerOrderNo, hungryCount, amount.Sub(totalAmount).String())
	//log.Info(msg)
	//dingding.ErrTransferDingBot.NotifyStr(msg)
	//go dao.ReportCollectAlert(int64(mchId), outerOrderNo, chain, coinType, amount.String())
	return false, errors.New(msg)
}

func getLessThanEnoughAmountList(mchId int, coinType, collectThreshold string, amount decimal.Decimal, pickCount int) ([]entity.FcAddressAmount, bool, decimal.Decimal, error) {
	var err error
	lessThanList, err := dao.FindLessThanOutAmountList(mchId, coinType, amount.String(), collectThreshold, pickCount)
	if err != nil {
		return nil, false, decimal.Zero, err
	}
	var pickedAddrs []entity.FcAddressAmount
	total := decimal.Zero
	enoughAmt := false
	for _, addrAmt := range lessThanList {
		amt, _ := decimal.NewFromString(addrAmt.Amount)
		total = total.Add(amt)
		pickedAddrs = append(pickedAddrs, addrAmt)

		// 这里表示如果足够出账，多拿一个地址
		if enoughAmt {
			break
		}

		if total.Cmp(amount) > 0 {
			enoughAmt = true
		}
	}
	return pickedAddrs, enoughAmt, total, nil
}

func coldAndUserAddrEnough(needAmt decimal.Decimal, mchId int, coinCode string, coldAddrAmt decimal.Decimal, count int) ([]entity.FcAddressAmount, bool, error) {
	userAddrAmountList, err := dao.FindUserAddrAmountList(mchId, coinCode, count)
	if err != nil {
		return nil, false, err
	}

	total := coldAddrAmt
	for _, uAddr := range userAddrAmountList {
		fs, _ := decimal.NewFromString(uAddr.Amount)
		total = total.Add(fs)
		if total.Cmp(needAmt) > 0 {
			return userAddrAmountList, true, nil
		}
	}
	return userAddrAmountList, false, nil
}

func doSendCollectCenter(url, outerOrderNo string, pickedAddrs []string, contract string, coinCode string, amount decimal.Decimal) error {
	taskReq := &CollectTaskRequest{
		OuterOrderNo: outerOrderNo,
		Address:      pickedAddrs,
		Contract:     contract,
		CoinCode:     coinCode,
		NeedAmount:   amount.String(),
	}
	resData, err := util.PostJson(url, taskReq)
	if err != nil {
		return fmt.Errorf("调用collect-center执行失败，返回结果: %v", err)
	}
	log.Infof("调用collect-center执行返回结果: %s", string(resData))
	result := &CollectTaskResult{}
	if err = json.Unmarshal(resData, result); err != nil {
		return fmt.Errorf("CollectTaskResult json.Unmarshal 失败: %v", err)
	}
	if result.Code != collectTaskSuccessCode {
		return fmt.Errorf("调用collect-center执行未成功: %s", result.Message)
	}
	return nil
}

func CallAdminCollectCenter(outerOrderNo string, pickedAddrs []string, contract string, coinCode string, amount decimal.Decimal) error {
	return doSendCollectCenter(fmt.Sprintf("%s/admin/collect/add", conf.Cfg.CollectCenter.Url), outerOrderNo, pickedAddrs, contract, coinCode, amount)
}

func CallCollectCenter(outerOrderNo string, pickedAddrs []string, contract string, coinCode string, amount decimal.Decimal) error {
	return doSendCollectCenter(fmt.Sprintf("%s/collect/add", conf.Cfg.CollectCenter.Url), outerOrderNo, pickedAddrs, contract, coinCode, amount)
}

func fetchAddress(addrAmounts []entity.FcAddressAmount) []string {
	var addressList []string
	for _, a := range addrAmounts {
		addressList = append(addressList, a.Address)
	}
	return addressList
}

func checkColdAddressAmountEnough(mchId int, chain, coinType string, amount decimal.Decimal) (bool, error) {
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   chain,
	})
	if err != nil {
		return false, err
	}

	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, amount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return false, fmt.Errorf("err:%s", err.Error())
	}
	return len(fromAddrs) > 0, nil
}

// getMaxAddressAmount 获取余额最多的出账地址
func getMaxAddressAmount(mchId int, chain, coinType string) (decimal.Decimal, error) {
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": mchId,
		"coin_name":   chain,
	})
	if err != nil {
		return decimal.Zero, err
	}

	fromAddrs := make([]entity.FcAddressAmount, 0)
	if err := db.Conn.Table("fc_address_amount").Where(builder.Expr("coin_type = ? and amount > 0 and forzen_amount = 0", coinType).
		And(builder.In("address", coldAddrs))).
		Find(&fromAddrs); err != nil {
		return decimal.Zero, err
	}

	//fromAddrs, err := entity.FcAddressAmount{}.Find(builder.Expr("coin_type = ? and amount > 0 and forzen_amount = 0", coinType).
	//	And(builder.In("address", coldAddrs)), 0)
	//if err != nil {
	//	return decimal.Zero, fmt.Errorf("err:%s", err.Error())
	//}

	max := decimal.Zero

	for _, addr := range fromAddrs {
		fs, _ := decimal.NewFromString(addr.Amount)
		if fs.Cmp(max) == 1 {
			max = fs
		}
	}
	return max, nil
}
