package deals

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/module/dict"
	"fmt"
)

func FindChainBillList(bill *domain.ChainBillSelect) ([]domain.ChainBillInfo, int64, error) {
	var (
		billList = []domain.ChainBillInfo{}
		dao      = chainBill.NewEntity()
	)
	list, count, err := dao.FindList(bill)
	if err != nil {
		return billList, count, err
	}

	for i, _ := range list {

		cf := ""
		if !list[i].ConfirmTime.IsZero() {
			cf = list[i].ConfirmTime.Local().Format(global.YyyyMmDdHhMmSs)
		}
		ct := ""
		if !list[i].CreatedAt.IsZero() {
			ct = list[i].CreatedAt.Local().Format(global.YyyyMmDdHhMmSs)
		}
		tx := ""
		if !list[i].TxTime.IsZero() {
			tx = list[i].TxTime.Local().Format(global.YyyyMmDdHhMmSs)
		}
		statusName := dict.BillStateList[list[i].BillStatus]
		billList = append(billList, domain.ChainBillInfo{
			Id:                   list[i].Id,
			TxId:                 list[i].TxId,
			SerialNo:             list[i].SerialNo,
			MerchantId:           list[i].MerchantId,
			Phone:                fmt.Sprintf("(%s)%s", list[i].PhoneCode, list[i].Phone),
			CoinId:               list[i].CoinId,
			ChainId:              list[i].ChainId,
			ServiceId:            list[i].ServiceId,
			CoinName:             list[i].CoinName,
			ChainName:            list[i].ChainName,
			ServiceName:          list[i].ServiceName,
			TxType:               list[i].TxType,
			BillStatus:           list[i].BillStatus,
			Nums:                 list[i].Nums,
			Fee:                  list[i].Fee,
			BurnFee:              list[i].BurnFee,
			DestroyFee:           list[i].DestroyFee,
			TxTypeName:           dict.TxTypeNameList[list[i].TxType],
			BillStatusName:       statusName,
			TxToAddr:             list[i].TxToAddr,
			TxFromAddr:           list[i].TxFromAddr,
			Remark:               list[i].Remark,
			Memo:                 list[i].Memo,
			State:                list[i].State,
			TxTime:               tx,
			Height:               list[i].Height,
			ConfirmNums:          list[i].ConfirmNums,
			IsWalletDeal:         list[i].IsWalletDeal,
			IsWalletDealName:     dict.BaseText[list[i].IsWalletDeal],
			IsColdWallet:         list[i].IsColdWallet,
			IsColdWalletName:     dict.BaseText[list[i].IsColdWallet],
			ColdWalletState:      list[i].ColdWalletState,
			ColdWalletStateName:  dict.WalletStateText[list[i].ColdWalletState],
			ColdWalletResult:     list[i].ColdWalletResult,
			ColdWalletResultName: dict.WalletResultText[list[i].ColdWalletResult],
			IsReback:             list[i].IsReback,
			IsRebackName:         dict.BaseText[list[i].IsReback],
			ColorType:            dict.BillStateColors[statusName],
			IsTest:               list[i].IsTest,
			IsTestName:           dict.IsTestText[list[i].IsTest],
			ConfirmTime:          cf,
			CreateTime:           ct,
			CreateByUser:         list[i].CreateByUser,
		})
	}
	return billList, count, err
}
