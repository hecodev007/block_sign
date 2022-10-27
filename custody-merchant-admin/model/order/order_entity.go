package order

import (
	"github.com/shopspring/decimal"
	"time"
)

type Orders struct {
	ChainId     int             `json:"chain_id" gorm:"column:chain_id"`
	CoinId      int             `json:"coin_id" gorm:"column:coin_id"`
	ServiceId   int             `json:"service_id" gorm:"column:service_id"`
	MerchantId  int64           `json:"merchant_id" gorm:"column:merchant_id"`
	Phone       string          `json:"phone" gorm:"column:phone"`
	ServiceName string          `json:"service_name" gorm:"column:service_name"`
	Type        int             `json:"type" gorm:"column:type"`
	AuditType   int             `json:"audit_type" gorm:"column:audit_type"`
	OrderResult int             `json:"order_result" gorm:"column:order_result"`
	AuditResult int             `json:"audit_result" gorm:"column:audit_result"`
	State       int             `json:"state" gorm:"column:state"`
	Id          int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	SerialNo    string          `gorm:"column:serial_no" json:"serial_no,omitempty"`
	TxId        string          `gorm:"column:tx_id" json:"tx_id,omitempty"`
	Memo        string          `json:"memo" gorm:"column:memo"`
	ReceiveAddr string          `json:"receive_addr" gorm:"column:receive_addr"`
	FromAddr    string          `json:"from_addr" gorm:"column:from_addr"`
	CoinName    string          `json:"coin_name" gorm:"column:coin_name"`
	ChainName   string          `json:"chain_name" gorm:"column:chain_name"`
	Reason      string          `json:"reason" gorm:"column:reason"`
	Nums        decimal.Decimal `json:"nums" gorm:"column:nums"`
	Fee         decimal.Decimal `json:"fee,omitempty"  gorm:"column:fee"`
	UpChainFee  decimal.Decimal `json:"up_chain_fee,omitempty"  gorm:"column:up_chain_fee"`
	BurnFee     decimal.Decimal `json:"burn_fee,omitempty"  gorm:"column:burn_fee"`
	DestroyFee  decimal.Decimal `json:"destroy_fee,omitempty"  gorm:"column:destroy_fee"`
	RealNums    decimal.Decimal `json:"real_nums,omitempty"  gorm:"column:real_nums"`
	CreateTime  time.Time       `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time       `json:"update_time" gorm:"column:update_time"`
	CreateUser  int64           `json:"create_user" gorm:"column:create_user"`
}

type ServiceOrderInfo struct {
	Nums int `json:"nums"  gorm:"column:nums"`
}

type WithdrawalOrderInfo struct {
	Nums   decimal.Decimal `json:"nums"  gorm:"column:nums"`
	Counts int             `json:"counts"  gorm:"column:counts"`
}

type ConfigNums struct {
	Nums decimal.Decimal `json:"nums"  gorm:"column:nums"`
}

type CountOrders struct {
	Count       int `json:"count"  gorm:"column:count"`
	OrderResult int `json:"order_result" gorm:"column:order_result"`
}

func (o *Orders) TableName() string {
	return "orders"
}
