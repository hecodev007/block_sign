package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"time"
)

type ColdAddrBalanceNotEnoughResult struct {
	ColdAddress      []ColdAddrBalanceNotEnoughInner `json:"coldAddress"`
	UserAddress      []ColdAddrBalanceNotEnoughInner `json:"userAddress"`
	ColdTotal        string                          `json:"coldTotal"`         // 出账地址总余额
	UserTotal        string                          `json:"userTotal"`         // 前30个用户地址总余额
	NeedAmount       string                          `json:"needAmount"`        // 出账金额
	CollectThreshold string                          `json:"collect_threshold"` // 币种的归集阈值
	LiquidBalance    string                          `json:"liquid_balance"`    // 币种的可用余额
}

type ColdAddrBalanceNotEnoughInner struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

func ReportColdAddrBalanceNotEnough(mchId int64, OuterOrderId, chain, coinType, amount string) {
	if err := coldAddrBalanceNotEnough(mchId, OuterOrderId, chain, coinType, amount, entity.ColdAddrBalanceNotEnough); err != nil {
		log.Infof("ReportColdAddrBalanceNotEnough 出错: %v", err)
	}
}

func coldAddrBalanceNotEnough(mchId int64, OuterOrderId, chain, coinType, amount string, reportType entity.ReportType) error {
	//m := &ColdAddrBalanceNotEnoughResult{NeedAmount: amount}
	//coldAddrList, err := FcFindAddressAmountFilter(mchId, coinType, address.AddressTypeCold, 100)
	//if err != nil {
	//	return err
	//}
	//coldTotal := decimal.Zero
	//for _, aa := range coldAddrList {
	//	m.ColdAddress = append(m.ColdAddress, ColdAddrBalanceNotEnoughInner{Address: aa.Address, Amount: aa.Amount})
	//	fs, _ := decimal.NewFromString(aa.Amount)
	//	coldTotal = coldTotal.Add(fs)
	//}
	//m.ColdTotal = coldTotal.String()
	//
	//userAddrList, err := FcFindAddressAmountFilter(mchId, coinType, address.AddressTypeUser, 30)
	//if err != nil {
	//	return err
	//}
	//userTotal := decimal.Zero
	//for _, aa := range userAddrList {
	//	m.UserAddress = append(m.UserAddress, ColdAddrBalanceNotEnoughInner{Address: aa.Address, Amount: aa.Amount})
	//	fs, _ := decimal.NewFromString(aa.Amount)
	//	userTotal = userTotal.Add(fs)
	//}
	//m.UserTotal = userTotal.String()
	//
	//coinSet, _ := FcCoinSetGetByName(coinType, 1)
	//if coinSet != nil {
	//	m.CollectThreshold = coinSet.CollectThreshold
	//}
	//
	//liquidByCoin, err := FcAddressAmountBalanceLiquidByCoin(mchId, coinType)
	//if err != nil {
	//	return err
	//}
	//m.LiquidBalance = liquidByCoin.Amount

	//ms, _ := json.Marshal(m)

	rr := &entity.FcReportRecord{
		Chain:        chain,
		CoinCode:     coinType,
		TxId:         "",
		ReportType:   reportType,
		OuterOrderId: OuterOrderId,
		Remark:       string("没有符合条件的出账地址"),
		CreateTime:   time.Now().Unix(),
	}

	return rr.Insert()
}

func ReportCollectAlert(mchId int64, OuterOrderId, chain, coinType, amount string) {
	if err := coldAddrBalanceNotEnough(mchId, OuterOrderId, chain, coinType, amount, entity.CollectAmountNotEnough); err != nil {
		log.Infof("ReportCollectAlert 出错: %v", err)
	}
}


func FcReportRecordGetByOuterOrderNo(outerOrderNo string) (*entity.FcReportRecord, error) {
	result := new(entity.FcReportRecord)
	has, err := db.Conn.Where("outer_order_id = ?", outerOrderNo).OrderBy("id desc").Limit(1).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return result, nil
}

func FcReportRecordUpdate(id int, remark string) error {
	model := &entity.FcReportRecord{Remark: remark, CreateTime: time.Now().Unix()}
	_, err := db.Conn.Where("id = ?", id).Update(model)
	if err != nil {
		return err
	}
	return nil
}
