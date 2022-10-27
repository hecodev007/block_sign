package dto

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/model/chainBill"
	"time"
)

func GetChainBillInfo(d domain.BillInfo) *chainBill.Entity {

	return &chainBill.Entity{
		TxId:             d.TxId,
		MerchantId:       d.MerchantId,
		Phone:            d.Phone,
		SerialNo:         d.SerialNo,
		TxToAddr:         d.TxToAddr,
		TxFromAddr:       d.TxFromAddr,
		TxType:           d.TxType,
		CoinId:           d.CoinId,
		ChainId:          d.ChainId,
		ServiceId:        d.ServiceId,
		BillStatus:       d.BillStatus,
		State:            d.State,
		Height:           0,
		ConfirmNums:      0,
		IsWalletDeal:     0,
		IsColdWallet:     1,
		ColdWalletState:  0,
		ColdWalletResult: 0,
		IsReback:         0,
		Remark:           d.Remark,
		Memo:             d.Memo,
		Nums:             d.Nums,
		Fee:              d.Fee,
		BurnFee:          d.BurnFee,
		DestroyFee:       d.DestroyFee,
		TxTime:           time.Now().Local(),
		ConfirmTime:      time.Now().Local(),
		CreateByUser:     d.CreateByUser,
		CreatedAt:        time.Now().Local(),
	}
}

func BillDetailToChainBill(d *bill.BillDetail) *chainBill.Entity {

	return &chainBill.Entity{
		TxId:             d.TxId,
		MerchantId:       d.MerchantId,
		Phone:            d.Phone,
		SerialNo:         d.SerialNo,
		TxToAddr:         d.TxToAddr,
		TxFromAddr:       d.TxFromAddr,
		TxType:           d.TxType,
		CoinId:           d.CoinId,
		ChainId:          d.ChainId,
		ServiceId:        d.ServiceId,
		BillStatus:       d.BillStatus,
		State:            d.State,
		Height:           0,
		ConfirmNums:      0,
		IsWalletDeal:     0,
		IsColdWallet:     1,
		ColdWalletState:  1,
		ColdWalletResult: 0,
		IsReback:         0,
		Remark:           d.Remark,
		Memo:             d.Memo,
		Nums:             d.RealNums, // 实际到账
		Fee:              d.Fee,
		BurnFee:          d.BurnFee,
		DestroyFee:       d.DestroyFee,
		TxTime:           time.Now().Local(),
		ConfirmTime:      time.Now().Local(),
		CreateByUser:     d.CreateByUser,
		CreatedAt:        time.Now().Local(),
	}
}
