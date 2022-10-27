package bill

import (
	"github.com/shopspring/decimal"
	"time"
)

type BillDetail struct {
	Id            int64           `gorm:"column:id; PRIMARY_KEY" json:"id"`
	TxId          string          `gorm:"column:tx_id" json:"tx_id,omitempty"`
	MerchantId    int64           `gorm:"column:merchant_id;" json:"merchant_id"`
	Phone         string          `gorm:"column:phone" json:"phone,omitempty"`
	SerialNo      string          `gorm:"column:serial_no" json:"serial_no,omitempty"`
	CoinName      string          `gorm:"column:coin_name" json:"coin_name,omitempty"`
	ChainName     string          `gorm:"column:chain_name" json:"chain_name,omitempty"`
	ServiceName   string          `gorm:"column:service_name" json:"service_name,omitempty"`
	TxToAddr      string          `gorm:"column:tx_to_addr" json:"tx_to_addr"`
	TxFromAddr    string          `gorm:"column:tx_from_addr" json:"tx_from_addr"`
	FromId        string          `gorm:"column:from_id" json:"from_id"`
	ToId          string          `gorm:"column:to_id" json:"to_id"`
	TxType        int             `gorm:"column:tx_type" json:"tx_type,omitempty"`
	CoinId        int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ChainId       int             `gorm:"column:chain_id" json:"chain_id"`
	ServiceId     int             `gorm:"column:service_id" json:"service_id,omitempty"`
	BillStatus    int             `gorm:"column:bill_status" json:"bill_status"`
	State         int             `gorm:"column:state" json:"state"`
	Remark        string          `gorm:"column:remark" json:"remark,omitempty"`
	Memo          string          `gorm:"column:memo" json:"memo,omitempty"`
	Nums          decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Fee           decimal.Decimal `json:"fee,omitempty"  gorm:"column:fee"`
	UpChainFee    decimal.Decimal `json:"up_chain_fee,omitempty"  gorm:"column:up_chain_fee"`
	BurnFee       decimal.Decimal `json:"burn_fee,omitempty"  gorm:"column:burn_fee"`
	DestroyFee    decimal.Decimal `json:"destroy_fee,omitempty"  gorm:"column:destroy_fee"`
	RealNums      decimal.Decimal `json:"real_nums,omitempty"  gorm:"column:real_nums"`
	WithdrawalFee decimal.Decimal `json:"withdrawal_fee,omitempty" gorm:"column:withdrawal_fee"`
	TopUpFee      decimal.Decimal `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"`
	CreatedAt     time.Time       `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time       `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt     *time.Time      `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
	TxTime        time.Time       `gorm:"column:tx_time" json:"tx_time,omitempty"`
	AuditTime     time.Time       `gorm:"column:audit_time" json:"audit_time,omitempty"`
	ConfirmTime   time.Time       `gorm:"column:confirm_time" json:"confirm_time,omitempty"`
	CreateByUser  int64           `gorm:"column:create_by_user" json:"create_by_user"`
}

type BillNums struct {
	Nums   decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	CoinId int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
}

type ConfigNums struct {
	Nums decimal.Decimal `json:"nums"  gorm:"column:nums"`
}
type BillLists struct {
	BillDetail
	OrderResult int `json:"order_result"  gorm:"column:order_result"`
}

type WithdrawalOrderInfo struct {
	Nums   decimal.Decimal `json:"nums"  gorm:"column:nums"`
	Counts int             `json:"counts"  gorm:"column:counts"`
}

func (b *BillDetail) TableName() string {
	return "bill_detail"
}
