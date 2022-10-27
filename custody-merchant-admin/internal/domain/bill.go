package domain

import (
	"github.com/shopspring/decimal"
)

type BillInfo struct {
	Id             int64           `json:"id,omitempty"`
	TxId           string          `json:"tx_id,omitempty"`
	SerialNo       string          `json:"serial_no,omitempty"`
	MerchantId     int64           `json:"merchant_id"`
	Phone          string          `json:"phone,omitempty"`
	CoinId         int             `json:"coin_id,omitempty"`
	ChainId        int             `json:"chin_id,omitempty"`
	ServiceId      int             `json:"service_id,omitempty"`
	CoinName       string          `json:"coin_name,omitempty"`
	ChainName      string          `json:"chain_name,omitempty"`
	ServiceName    string          `json:"service_name,omitempty"`
	TxType         int             `json:"tx_type,omitempty"`
	BillStatus     int             `json:"bill_status"`
	FromId         string          `json:"from_id"`
	ToId           string          `json:"to_id"`
	Nums           decimal.Decimal `json:"nums,omitempty"`
	Price          decimal.Decimal `json:"price,omitempty"`
	Fee            decimal.Decimal `json:"fee,omitempty"`
	UpChainFee     decimal.Decimal `json:"up_chain_fee,omitempty"`
	BurnFee        decimal.Decimal `json:"burn_fee,omitempty"`
	DestroyFee     decimal.Decimal `json:"destroy_fee,omitempty"`
	RealNums       decimal.Decimal `json:"real_nums,omitempty"`
	WithdrawalFee  decimal.Decimal `json:"withdrawal_fee,omitempty"`
	TopUpFee       decimal.Decimal `json:"top_up_fee,omitempty"`
	TxTypeName     string          `json:"tx_type_name,omitempty"`
	BillStatusName string          `json:"bill_status_name"`
	TxToAddr       string          `json:"tx_to_addr"`
	TxFromAddr     string          `json:"tx_from_addr"`
	Remark         string          `json:"remark,omitempty"`
	Memo           string          `json:"memo,omitempty"`
	State          int             `json:"state"`
	OrderResult    int             `json:"order_result"`
	ResultName     string          `json:"result_name"`
	TxTime         string          `json:"tx_time,omitempty"`
	AuditTime      string          `json:"audit_time,omitempty"`
	ConfirmTime    string          `json:"confirm_time,omitempty"`
	CreateTime     string          `json:"create_time,omitempty"`
	CreateByUser   int64           `json:"create_by_user"`
	ColorType      string          `json:"color_type"`
	ColorResult    string          `json:"color_result"`
}

type BillSelect struct {
	MerchantId       int64    `json:"merchant_id"`
	Phone            string   `json:"phone,omitempty"`
	CoinId           int      `json:"coin_id,omitempty"`
	ServiceId        int      `json:"service_id,omitempty"`
	TxType           int      `json:"tx_type,omitempty"`
	BillStatus       int      `json:"bill_status"`
	Title            []string `json:"title"`
	Limit            int      `json:"limit"  description:"查询条数" example:"10"`
	Offset           int      `json:"offset" description:"查询起始位置" example:"0"`
	CreateByUser     int64    `json:"create_by_user"`
	UnitId           int      `json:"unit_id"`
	Min              float64  `json:"min"`
	Max              float64  `json:"max"`
	Address          string   `json:"address,omitempty"`
	TxStartTime      string   `json:"tx_start_time"`
	TxEndTime        string   `json:"tx_end_time"`
	ConfirmStartTime string   `json:"confirm_start_time"`
	ConfirmEndTime   string   `json:"confirm_end_time"`
}

type BillBalance struct {
	UnitId      int           `json:"unit_id"`
	BalanceList []BalanceList `json:"balance_list"`
}

type BalanceList struct {
	Title    string          `json:"title"`
	Icon     string          `json:"icon"`
	UsdtNums decimal.Decimal `json:"usdt_nums"`
	UnitNums string          `json:"unit_nums"`
}
