package chainBill

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

type Entity struct {
	Db               *orm.CacheDB    `json:"-" gorm:"-"`
	Id               int64           `gorm:"column:id; PRIMARY_KEY" json:"id"`
	TxId             string          `gorm:"column:tx_id" json:"tx_id,omitempty"`
	MerchantId       int64           `gorm:"column:merchant_id;" json:"merchant_id"`
	Phone            string          `gorm:"column:phone" json:"phone,omitempty"`
	SerialNo         string          `gorm:"column:serial_no" json:"serial_no,omitempty"`
	TxToAddr         string          `gorm:"column:tx_to_addr" json:"tx_to_addr"`
	TxFromAddr       string          `gorm:"column:tx_from_addr" json:"tx_from_addr"`
	TxType           int             `gorm:"column:tx_type" json:"tx_type,omitempty"`
	CoinId           int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ChainId          int             `gorm:"column:chain_id" json:"chain_id,omitempty"`
	ServiceId        int             `gorm:"column:service_id" json:"service_id,omitempty"`
	BillStatus       int             `gorm:"column:bill_status" json:"bill_status"`
	State            int             `gorm:"column:state" json:"state"`
	Height           int             `gorm:"column:height" json:"height"`
	ConfirmNums      int             `gorm:"column:confirm_nums" json:"confirm_nums"`
	IsWalletDeal     int             `gorm:"column:is_wallet_deal" json:"is_wallet_deal"`
	IsColdWallet     int             `gorm:"column:is_cold_wallet" json:"is_cold_wallet"`
	ColdWalletState  int             `gorm:"column:cold_wallet_state" json:"cold_wallet_state"`
	ColdWalletResult int             `gorm:"column:cold_wallet_result" json:"cold_wallet_result"`
	IsReback         int             `gorm:"column:is_reback" json:"is_reback"`
	Remark           string          `gorm:"column:remark" json:"remark,omitempty"`
	Memo             string          `gorm:"column:memo" json:"memo,omitempty"`
	Nums             decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Fee              decimal.Decimal `json:"fee,omitempty"  gorm:"column:fee"`
	UpChainFee       decimal.Decimal `json:"up_chain_fee,omitempty"  gorm:"column:up_chain_fee"`
	BurnFee          decimal.Decimal `json:"burn_fee,omitempty"  gorm:"column:burn_fee"`
	DestroyFee       decimal.Decimal `json:"destroy_fee,omitempty"  gorm:"column:destroy_fee"`
	TxTime           time.Time       `gorm:"column:tx_time" json:"tx_time,omitempty"`
	ConfirmTime      time.Time       `gorm:"column:confirm_time" json:"confirm_time,omitempty"`
	CreateByUser     int64           `gorm:"column:create_by_user" json:"create_by_user"`
	CreatedAt        time.Time       `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt        time.Time       `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt        time.Time       `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (e *Entity) TableName() string {
	return "chain_bill"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type ChainBillLists struct {
	Id               int64           `gorm:"column:id; PRIMARY_KEY" json:"id"`
	TxId             string          `gorm:"column:tx_id" json:"tx_id,omitempty"`
	MerchantId       int64           `gorm:"column:merchant_id;" json:"merchant_id"`
	PhoneCode        string          `gorm:"column:phone_code" json:"phone_code,omitempty"`
	Phone            string          `gorm:"column:phone" json:"phone,omitempty"`
	SerialNo         string          `gorm:"column:serial_no" json:"serial_no,omitempty"`
	TxToAddr         string          `gorm:"column:tx_to_addr" json:"tx_to_addr"`
	TxFromAddr       string          `gorm:"column:tx_from_addr" json:"tx_from_addr"`
	TxType           int             `gorm:"column:tx_type" json:"tx_type,omitempty"`
	CoinId           int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ChainId          int             `gorm:"column:chain_id" json:"chain_id,omitempty"`
	ServiceId        int             `gorm:"column:service_id" json:"service_id,omitempty"`
	BillStatus       int             `gorm:"column:bill_status" json:"bill_status"`
	State            int             `gorm:"column:state" json:"state"`
	Height           int             `gorm:"column:height" json:"height"`
	ConfirmNums      int             `gorm:"column:confirm_nums" json:"confirm_nums"`
	IsWalletDeal     int             `gorm:"column:is_wallet_deal" json:"is_wallet_deal"`
	IsColdWallet     int             `gorm:"column:is_cold_wallet" json:"is_cold_wallet"`
	ColdWalletState  int             `gorm:"column:cold_wallet_state" json:"cold_wallet_state"`
	ColdWalletResult int             `gorm:"column:cold_wallet_result" json:"cold_wallet_result"`
	IsReback         int             `gorm:"column:is_reback" json:"is_reback"`
	Remark           string          `gorm:"column:remark" json:"remark,omitempty"`
	Memo             string          `gorm:"column:memo" json:"memo,omitempty"`
	Nums             decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Fee              decimal.Decimal `json:"fee,omitempty"  gorm:"column:fee"`
	BurnFee          decimal.Decimal `json:"burn_fee,omitempty"  gorm:"column:burn_fee"`
	DestroyFee       decimal.Decimal `json:"destroy_fee,omitempty"  gorm:"column:destroy_fee"`
	TxTime           time.Time       `gorm:"column:tx_time" json:"tx_time,omitempty"`
	ConfirmTime      time.Time       `gorm:"column:confirm_time" json:"confirm_time,omitempty"`
	CreateByUser     int64           `gorm:"column:create_by_user" json:"create_by_user"`
	CreatedAt        time.Time       `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt        time.Time       `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt        time.Time       `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
	CoinName         string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	ChainName        string          `json:"chain_name,omitempty" gorm:"column:chain_name"`
	ServiceName      string          `json:"service_name,omitempty" gorm:"column:service_name"`
	IsTest           int             `json:"is_test" gorm:"column:is_test"`
}
