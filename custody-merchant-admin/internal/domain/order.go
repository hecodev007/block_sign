package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type OrderInfo struct {
	CoinId      int             `json:"coin_id"`
	ChainId     int             `json:"chain_id"`
	ServiceId   int             `json:"service_id"`
	MerchantId  int64           `json:"merchant_id"`
	Phone       string          `json:"phone"`
	Type        int             `json:"type"`
	Id          int64           `json:"id"`
	SerialNo    string          `json:"serial_no,omitempty"`
	TxId        string          `json:"tx_id,omitempty"`
	Memo        string          `json:"memo"`
	ReceiveAddr string          `json:"receive_addr"`
	FromAddr    string          `json:"from_addr"`
	Nums        decimal.Decimal `json:"nums"`
	Fee         decimal.Decimal `json:"fee,omitempty"`
	UpChainFee  decimal.Decimal `json:"up_chain_fee,omitempty"`
	BurnFee     decimal.Decimal `json:"burn_fee,omitempty"`
	DestroyFee  decimal.Decimal `json:"destroy_fee,omitempty"`
	RealNums    decimal.Decimal `json:"real_nums,omitempty"`
	CreateUser  int64           `json:"create_user,omitempty"`
}

type SelectOrderInfo struct {
	ServiceId   int    `json:"service_id"`
	CoinId      int    `json:"coin_id"`
	OrderResult int    `json:"order_result"`
	SerialNo    string `json:"serial_no"`
	ChainName   string `json:"chain_name"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Contents    string `json:"contents"`
}

type SelectOrderList struct {
	CoinId        int             `json:"coin_id"`
	ChainId       int             `json:"chain_id"`
	Phone         string          `json:"phone"`
	MerchantId    int64           `json:"merchant_id"`
	ServiceId     int             `json:"service_id"`
	ServiceName   string          `json:"service_name"`
	SerialNo      string          `json:"serial_no,omitempty"`
	Type          int             `json:"type"`
	OrderResult   int             `json:"order_result"`
	AuditStatus   int             `json:"audit_status"`
	Id            int64           `json:"id"`
	Memo          string          `json:"memo"`
	ReceiveAddr   string          `json:"receive_addr"`
	CoinName      string          `json:"coin_name"`
	ChainName     string          `json:"chain_name"`
	TypeName      string          `json:"type_name"`
	ResultName    string          `json:"result_name"`
	AuditType     int             `json:"audit_type"`
	AuditTypeName string          `json:"audit_type_name"`
	Nums          decimal.Decimal `json:"nums"`
	NumsPrice     decimal.Decimal `json:"nums_price"`
	Fee           decimal.Decimal `json:"fee"`
	UpChainFee    decimal.Decimal `json:"up_chain_fee,omitempty"`
	BurnFee       decimal.Decimal `json:"burn_fee,omitempty"`
	DestroyFee    decimal.Decimal `json:"destroy_fee,omitempty"`
	RealNums      decimal.Decimal `json:"real_nums,omitempty"`
	RealNumsPrice decimal.Decimal `json:"real_nums_price,omitempty"`
	CreateTime    string          `json:"create_time"`
	AuditTime     string          `json:"audit_time"`
	AuditNames    string          `json:"audit_names"`
	Reason        string          `json:"reason"`
	ColorResult   string          `json:"color_result"`
}

type UpdateOrders struct {
	Id          int64  `json:"id"`
	AuditStatus int    `json:"audit_status"`
	Reason      string `json:"reason"`
}

type CountStatus struct {
	Count     int    `json:"count"`
	StateName string `json:"state_name"`
	State     int    `json:"state"`
}
type CountLevel struct {
	Count      int `json:"count"`
	AuditLevel int `json:"audit_level"`
}

type AuditDetail struct {
	State      int    `json:"state"`
	AuditLevel int    `json:"audit_level"`
	AuditType  int    `json:"audit_type"`
	TypeName   string `json:"type_name"`
	StateName  string `json:"state_name"`
	LevelName  string `json:"level_name"`
	UserId     int64  `json:"user_id"`
	UserName   string `json:"user_name"`
	StatusName string `json:"status_name"`
	UpDateTime string `json:"update_time,omitempty"`
}

type OrderDetail struct {
	State      int         `json:"state"`
	AuditType  int         `json:"audit_type"`
	Icon       string      `json:"icon"`
	TypeName   string      `json:"type_name"`
	StatusName string      `json:"status_name"`
	UserList   []UserAudit `json:"user_list"`
	UpDateTime string      `json:"update_time,omitempty"`
}

type UserAudit struct {
	AuditResult int    `json:"audit_result"`
	ResultName  string `json:"result_name"`
	AuditLevel  int    `json:"audit_level"`
	LevelName   string `json:"level_name"`
	UserId      int64  `json:"user_id"`
	UserName    string `json:"user_name"`
}

type OrderHistory struct {
	OrderId    int64     `json:"order_id"`
	SelectTime time.Time `json:"select_time"`
}
