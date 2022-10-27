package domain

import (
	"github.com/shopspring/decimal"
)

type ChainBillInfo struct {
	Id                   int64           `json:"id,omitempty"`
	TxId                 string          `json:"tx_id,omitempty"`
	SerialNo             string          `json:"serial_no,omitempty"`
	MerchantId           int64           `json:"merchant_id"`
	Phone                string          `json:"phone,omitempty"`
	CoinId               int             `json:"coin_id,omitempty"`
	ChainId              int             `json:"chin_id,omitempty"`
	ServiceId            int             `json:"service_id,omitempty"`
	CoinName             string          `json:"coin_name,omitempty"`
	ChainName            string          `json:"chain_name,omitempty"`
	ServiceName          string          `json:"service_name,omitempty"`
	TxType               int             `json:"tx_type,omitempty"`
	BillStatus           int             `json:"bill_status"`
	Nums                 decimal.Decimal `json:"nums,omitempty"`
	Fee                  decimal.Decimal `json:"fee,omitempty"`
	BurnFee              decimal.Decimal `json:"burn_fee,omitempty"`
	DestroyFee           decimal.Decimal `json:"destroy_fee,omitempty"`
	TxTypeName           string          `json:"tx_type_name,omitempty"`
	BillStatusName       string          `json:"bill_status_name"`
	TxToAddr             string          `json:"tx_to_addr"`
	TxFromAddr           string          `json:"tx_from_addr"`
	Remark               string          `json:"remark,omitempty"`
	Memo                 string          `json:"memo,omitempty"`
	State                int             `json:"state"`
	Height               int             `json:"height"`
	ConfirmNums          int             `json:"confirm_nums"`
	IsWalletDeal         int             `json:"is_wallet_deal"`
	IsColdWallet         int             `json:"is_cold_wallet"`
	ColdWalletState      int             `json:"cold_wallet_state"`
	ColdWalletResult     int             `json:"cold_wallet_result"`
	IsWalletDealName     string          `json:"is_wallet_deal_name"`
	IsColdWalletName     string          `json:"is_cold_wallet_name"`
	ColdWalletStateName  string          `json:"wallet_state_name"`
	ColdWalletResultName string          `json:"wallet_result_name"`
	IsReback             int             `json:"is_reback"`
	IsRebackName         string          `json:"is_reback_name"`
	IsTest               int             `json:"is_test"`
	IsTestName           string          `json:"is_test_name"`
	TxTime               string          `json:"tx_time,omitempty"`
	CreateTime           string          `json:"create_time,omitempty"`
	ConfirmTime          string          `json:"confirm_time,omitempty"`
	CreateByUser         int64           `json:"create_by_user"`
	ColorType            string          `json:"color_type"`
}

type ChainBillSelect struct {
	MerchantId       int64  `json:"merchant_id"`
	Phone            string `json:"phone,omitempty"`
	TxType           int    `json:"tx_type,omitempty"`
	Limit            int    `json:"limit"  description:"查询条数" example:"10"`
	Offset           int    `json:"offset" description:"查询起始位置" example:"0"`
	AddressOrMemo    string `json:"address_or_memo"`
	IsReback         int    `json:"is_reback"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	ConfirmStartTime string `json:"confirm_start_time"`
	ConfirmEndTime   string `json:"confirm_end_time"`
}

type UpChainInfo struct {
	Id int64 `json:"id"`
}
