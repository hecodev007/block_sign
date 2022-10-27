package domain

import (
	"github.com/shopspring/decimal"
)

type BusinessReqInfo struct {
	Id         int64  `json:"id" query:"id"`
	ContactStr string `json:"contact_str" query:"contact_str"`
	AccountId  int64  `json:"account_id" query:"account_id"`
	BusinessId int64  `json:"business_id" query:"business_id"`
	Limit      int    `json:"limit"  query:"limit" description:"查询条数" example:"10"`
	Offset     int    `json:"offset" query:"offset" description:"查询起始位置" example:"0"`
}

type BusinessOperateInfo struct {
	Id      int64  `json:"id" query:"id"`
	Operate string `json:"operate" query:"operate"` //agree-同意，refuse-拒绝,lock-冻结，unlock-解冻
	Remark  string `json:"remark" query:"remark"`   //审核备注
}

////BusinessSecurityReq 请求安全信息
//type BusinessSecurityReq struct {
//	AccountId int64  `json:"account_id" form:"account_id"`
//	TypeName  string `json:"type_name" form:"type_name"` //clientid-客户id，secret-secret
//}

type BusinessInfo struct {
	PackageInfo
	Id              int64           `json:"id" form:"id"`
	AccountId       int64           `json:"account_id" form:"account_id"`
	AccountStatus   int             `json:"account_status" form:"account_status"`
	BusinessStatus  int             `json:"business_status" form:"business_status"`
	Phone           string          `json:"phone" form:"phone"`
	TradeType       string          `json:"trade_type,omitempty" gorm:"column:trade_type"`         //交易类型
	DeductCoinId    int             `json:"deduct_coin_id,omitempty" gorm:"column:deduct_coin_id"` //扣费币种id
	DeductCoin      string          `json:"deduct_coin,omitempty" gorm:"column:deduct_coin"`       //扣费币种
	BusinessName    string          `json:"business_name" form:"business_name"`
	BusinessId      int             `json:"business_id" form:"business_id"`
	PackageId       int             `json:"package_id" form:"package_id"`
	Coin            string          `json:"coin" form:"coin"`
	SubCoin         string          `json:"sub_coin" form:"sub_coin"`
	MoreCoinFee     decimal.Decimal `json:"more_coin_fee" form:"more_coin_fee"`
	MoreSubcoinFee  decimal.Decimal `json:"more_subcoin_fee" form:"more_subcoin_fee"`
	IsPlatformCheck int             `json:"is_platform_check" form:"is_platform_check"`
	IsAccountCheck  int             `json:"is_account_check" form:"is_account_check"`
	ClientId        string          `json:"client_id" form:"client_id"`
	Secret          string          `json:"secret" form:"secret"`
	IpAddr          string          `json:"ip_addr" form:"ip_addr"`
	CallbackUrl     string          `json:"callback_url" form:"callback_url"`
	IsSms           int             `json:"is_sms" form:"is_sms"`
	IsEmail         int             `json:"is_email" form:"is_email"`
	IsWithdrawal    int             `json:"is_withdrawal" form:"is_withdrawal"`
	IsIpVerify      int             `json:"is_ip_verify" form:"is_ip_verify"`
}

type BusinessListInfo struct {
	//Id             int64           `json:"id" form:"id"`
	AccountId      int64           `json:"account_id" form:"account_id"`
	Name           string          `json:"name" gorm:"column:name"`
	Email          string          `json:"email" gorm:"column:email"`
	Phone          string          `json:"phone" gorm:"column:phone"`
	IsTest         int             `json:"is_test" form:"is_test"`
	BusinessName   string          `json:"business_name" form:"business_name"`
	BusinessId     int             `json:"business_id" form:"business_id"`
	CreateTime     string          `json:"create_time" gorm:"create_time"`
	Coin           string          `json:"coin" gorm:"coin"`                  //主链币
	SubCoin        string          `json:"sub_coin" gorm:"sub_coin"`          //代币
	TypeName       string          `json:"type_name" gorm:"column:type_name"` //套餐类型
	ModelName      string          `json:"model_name" gorm:"column:model_name"`
	ProfitNumber   decimal.Decimal `json:"profit_number" gorm:"profit_number"` //套餐获益户
	OrderType      string          `json:"order_type" gorm:"order_type"`       //交易类型
	TopUpType      int             `json:"top_up_type" gorm:"column:top_up_type"`
	TopUpFee       string          `json:"top_up_fee" gorm:"column:top_up_fee"`
	WithdrawalType int             `json:"withdrawal_type" gorm:"column:withdrawal_type"`
	WithdrawalFee  string          `json:"withdrawal_fee" gorm:"column:withdrawal_fee"`
	CheckerName    string          `json:"checker_name" gorm:"checker_name"`
	BusinessStatus int             `json:"business_status" form:"business_status"`
	CheckedAt      string          `json:"checked_at" gorm:"checked_at"`
	Remark         string          `json:"remark" form:"remark"` //审核备注

}

type CreateBusinessInfo struct {
	PackageInfo
	AccountId       int64  `json:"account_id" form:"account_id" validate:"required"`
	Phone           string `json:"phone" form:"phone" validate:"required"`
	Email           string `json:"email" form:"email" validate:"required"`
	TradeType       string `json:"trade_type,omitempty" gorm:"column:trade_type" validate:"required"`             //交易类型
	DeductCoinId    string `json:"deduct_coin_id,omitempty" gorm:"column:deduct_coin_id"`                         //扣费币种id
	DeductCoinName  string `json:"deduct_coin_name,omitempty" gorm:"column:deduct_coin_name" validate:"required"` //扣费币种
	BusinessName    string `json:"business_name" form:"business_name" validate:"required"`
	PackageId       int    `json:"package_id" form:"package_id" validate:"required"`
	Coin            string `json:"coin" form:"coin" validate:"required"`
	SubCoin         string `json:"sub_coin" form:"sub_coin" validate:"required"`
	IsPlatformCheck int    `json:"is_platform_check" form:"is_platform_check"`
	IsAccountCheck  int    `json:"is_account_check" form:"is_account_check"`
	IpAddr          string `json:"ip_addr" form:"ip_addr"`
	CallbackUrl     string `json:"callback_url" form:"callback_url"`
	IsSms           int    `json:"is_sms" form:"is_sms"`
	IsEmail         int    `json:"is_email" form:"is_email"`
	IsWithdrawal    int    `json:"is_withdrawal" form:"is_withdrawal"`
	IsWhitelist     int    `json:"is_whitelist" form:"is_whitelist"`
	IsIp            int    `json:"is_ip" form:"is_ip"`
}

type BusinessDetailInfo struct {
	PackageInfo
	AccountId       int64  `json:"account_id" form:"account_id" validate:"required"`
	AccountStatus   int    `json:"account_status" gorm:"account_status"`
	Phone           string `json:"phone" form:"phone" validate:"required"`
	Email           string `json:"email" form:"email" validate:"required"`
	TradeType       string `json:"trade_type,omitempty" gorm:"column:trade_type" validate:"required"`         //交易类型
	DeductCoinId    int    `json:"deduct_coin_id,omitempty" gorm:"column:deduct_coin_id" validate:"required"` //扣费币种id
	DeductCoinName  string `json:"deduct_coin_name" gorm:"column:deduct_coin_name" validate:"required"`       //扣费币种
	DeductCoin      string `json:"deduct_coin" gorm:"column:deduct_coin"`                                     //扣费币种
	BusinessName    string `json:"business_name" form:"business_name" validate:"required"`
	PackageId       int    `json:"package_id" form:"package_id" validate:"required"`
	Coin            string `json:"coin" form:"coin" validate:"required"`
	SubCoin         string `json:"sub_coin" form:"sub_coin" validate:"required"`
	IsPlatformCheck int    `json:"is_platform_check" form:"is_platform_check"`
	IsAccountCheck  int    `json:"is_account_check" form:"is_account_check"`
	IpAddr          string `json:"ip_addr" form:"ip_addr"`
	CallbackUrl     string `json:"callback_url" form:"callback_url"`
	IsSms           int    `json:"is_sms" form:"is_sms"`
	IsEmail         int    `json:"is_email" form:"is_email"`
	IsWithdrawal    int    `json:"is_withdrawal" form:"is_withdrawal"`
	IsWhitelist     int    `json:"is_whitelist" form:"is_whitelist"`
	IsIp            int    `json:"is_ip" form:"is_ip"`
	ClientId        string `json:"client_id" form:"client_id"`
	ApiSecret       string `json:"api_secret" form:"api_secret"`
}

type BusinessPackageInfo struct {
	Id                int64           `json:"id" form:"id"`
	TypeName          string          `json:"type_name" form:"type_name"`
	ModelName         string          `json:"model_name" form:"model_name"`
	OrderType         string          `json:"order_type" form:"order_type"`
	TradeType         string          `json:"trade_type" form:"column:trade_type"`
	DeployFee         decimal.Decimal `json:"deploy_fee" form:"deploy_fee"`
	CustodyFee        decimal.Decimal `json:"custody_fee" form:"custody_fee"`
	AddrNums          int             `json:"addr_nums" form:"addr_nums"`
	DepositFee        decimal.Decimal `json:"deposit_fee" form:"deposit_fee"`
	CoverFee          decimal.Decimal `json:"cover_fee" form:"cover_fee"`
	MoreCoinFee       decimal.Decimal `json:"more_coin_fee" form:"more_coin_fee"`
	MoreSubcoinFee    decimal.Decimal `json:"more_subcoin_fee" form:"more_subcoin_fee"`
	ComboDiscountUnit int             `json:"combo_discount_unit" form:"combo_discount_unit"`
	ComboDiscount     decimal.Decimal `json:"combo_discount" form:"combo_discount"`
	ComboDiscountNums decimal.Decimal `json:"combo_discount_nums" form:"combo_discount_nums"`
	YearDiscountUnit  int             `json:"year_discount_unit" form:"year_discount_unit"`
	YearDiscount      decimal.Decimal `json:"year_discount" form:"year_discount"`
	YearDiscountNums  decimal.Decimal `json:"year_discount_nums" form:"year_discount_nums"`
}

type BusinessSecurityInfo struct {
	ClientId     string `json:"client_id" form:"client_id"`
	Secret       string `json:"secret" form:"secret"`
	IpAddr       string `json:"ip_addr" form:"ip_addr"`
	CallbackUrl  string `json:"callback_url" form:"callback_url"`
	IsWithdrawal int    `json:"is_withdrawal" form:"is_withdrawal"`
	IsIp         int    `json:"is_ip" form:"is_ip"`
	Phone        string `json:"phone" form:"phone"`
	Email        string `json:"email" form:"email"`
}

type BusinessSecurityBoolInfo struct {
	ClientId     string `json:"client_id" form:"client_id"`
	Secret       string `json:"secret" form:"secret"`
	IpAddr       string `json:"ip_addr" form:"ip_addr"`
	CallbackUrl  string `json:"callback_url" form:"callback_url"`
	IsWithdrawal bool   `json:"is_withdrawal" form:"is_withdrawal"`
	IsIp         bool   `json:"is_ip" form:"is_ip"`
	Phone        string `json:"phone" form:"phone"`
	Email        string `json:"email" form:"email"`
}

type CoinReqInfo struct {
	Name string `json:"name" query:"name"`
}

type OrderOperateReq struct {
	OutOrderId string `json:"out_order_id" query:"out_order_id"`
	BusinessId int64  `json:"business_id" query:"business_id"`
	AccountId  int    `json:"account_id" query:"account_id"`
}
