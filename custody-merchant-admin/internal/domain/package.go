package domain

import "github.com/shopspring/decimal"

//
//type ListBase struct {
//}

type PackageReqInfo struct {
	Id        int64  `json:"id" query:"id"`
	Screen    string `json:"screen" query:"screen"`
	TypeName  string `json:"type_name" query:"type_name"`
	ModelName string `json:"model_name" query:"model_name"`
	Limit     int    `json:"limit"   query:"limit" description:"查询条数" example:"10"`
	Offset    int    `json:"offset"  query:"offset" description:"查询起始位置" example:"0"`
}
type MchPackageReqInfo struct {
	PackageId int64 `json:"package_id" query:"package_id"`
	AccountId int64 `json:"account_id" query:"account_id"`
}

type PackageInfo struct {
	Id                  int64           `json:"id" form:"id"`
	TypeName            string          `json:"type_name" form:"type_name"`
	ModelName           string          `json:"model_name" form:"model_name"`
	EnterUnit           int             `json:"enter_unit" form:"enter_unit"` //入资单位 1-usdt,2-地址数量
	LimitType           int             `json:"limit_type" form:"limit_type"` //1：>,2 <,3: =
	TypeNums            decimal.Decimal `json:"type_nums" form:"type_nums"`
	TopUpType           int             `json:"top_up_type" form:"top_up_type"`
	TopUpFee            string          `json:"top_up_fee" form:"top_up_fee"` //充值收费 百分数/usdt，带单位
	WithdrawalType      int             `json:"withdrawal_type" form:"withdrawal_type"`
	WithdrawalFee       string          `json:"withdrawal_fee" form:"withdrawal_fee"` //提现收费类型 百分数/usdt，带单位
	ServiceNums         int             `json:"service_nums" form:"service_nums"`
	ServiceDiscountUnit int             `json:"service_discount_unit" form:"service_discount_unit"`
	ServiceDiscountNums decimal.Decimal `json:"service_discount_nums" form:"service_discount_nums"`
	ChainNums           int             `json:"chain_nums" form:"chain_nums"`
	ChainDiscountUnit   int             `json:"chain_discount_unit" form:"chain_discount_unit"`
	ChainDiscountNums   decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	ChainTimeUnit       int             `json:"chain_time_unit" form:"chain_time_unit"`
	CoinNums            int             `json:"coin_nums" form:"coin_nums"`
	CoinDiscountUnit    int             `json:"coin_discount_unit" form:"coin_discount_unit"`
	CoinDiscountNums    decimal.Decimal `json:"coin_discount_nums" form:"coin_discount_nums"`
	CoinTimeUnit        int             `json:"coin_time_unit" form:"coin_time_unit"`
	DeployFee           decimal.Decimal `json:"deploy_fee" form:"deploy_fee"`
	CustodyFee          decimal.Decimal `json:"custody_fee" form:"custody_fee"`
	DepositFee          decimal.Decimal `json:"deposit_fee" form:"deposit_fee"`
	AddrNums            int             `json:"addr_nums" form:"addr_nums"`
	CoverFee            decimal.Decimal `json:"cover_fee" form:"cover_fee"`
	ComboDiscountUnit   int             `json:"combo_discount_unit" form:"combo_discount_unit"`
	ComboDiscountNums   decimal.Decimal `json:"combo_discount_nums" form:"combo_discount_nums"`
	YearDiscountUnit    int             `json:"year_discount_unit" form:"year_discount_unit"`
	YearDiscountNums    decimal.Decimal `json:"year_discount_nums" form:"year_discount_nums"`
	IsPlatformCheck     int             `json:"is_platform_check" gorm:"is_platform_check"`
	IsAccountCheck      int             `json:"is_account_check" gorm:"is_account_check"`
}

type PackageListInfo struct {
	Id                  int64           `json:"id" query:"id"`
	TypeName            string          `json:"type_name" form:"type_name"`
	ModelName           string          `json:"model_name" form:"model_name"`
	EnterUnit           int             `json:"enter_unit" form:"enter_unit"` //入资单位 1-usdt,2-地址数量
	LimitType           int             `json:"limit_type" form:"limit_type"` //1：>,2 <,3: =
	TypeNums            decimal.Decimal `json:"type_nums" form:"type_nums"`
	TopUpType           int             `json:"top_up_type" form:"top_up_type"`
	TopUpFee            string          `json:"top_up_fee" form:"top_up_fee"` //充值收费 百分数/usdt，带单位
	WithdrawalType      int             `json:"withdrawal_type" form:"withdrawal_type"`
	WithdrawalFee       string          `json:"withdrawal_fee" form:"withdrawal_fee"` //提现收费类型 百分数/usdt，带单位
	ServiceNums         int             `json:"service_nums" form:"service_nums"`
	ServiceDiscountUnit int             `json:"service_discount_unit" form:"service_discount_unit"`
	ServiceDiscountNums decimal.Decimal `json:"service_discount_nums" form:"service_discount_nums"`
	ChainNums           int             `json:"chain_nums" form:"chain_nums"`
	ChainDiscountUnit   int             `json:"chain_discount_unit" form:"chain_discount_unit"`
	ChainDiscountNums   decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	ChainTimeUnit       int             `json:"chain_time_unit" form:"chain_time_unit"`
	CoinNums            int             `json:"coin_nums" form:"coin_nums"`
	CoinDiscountUnit    int             `json:"coin_discount_unit" form:"coin_discount_unit"`
	CoinDiscountNums    decimal.Decimal `json:"coin_discount_nums" form:"coin_discount_nums"`
	CoinTimeUnit        int             `json:"coin_time_unit" form:"coin_time_unit"`
	DeployFee           decimal.Decimal `json:"deploy_fee" form:"deploy_fee"`
	CustodyFee          decimal.Decimal `json:"custody_fee" form:"custody_fee"`
	DepositFee          decimal.Decimal `json:"deposit_fee" form:"deposit_fee"`
	AddrNums            int             `json:"addr_nums" form:"addr_nums"`
	CoverFee            decimal.Decimal `json:"cover_fee" form:"cover_fee"`
	ComboDiscountUnit   int             `json:"combo_discount_unit" form:"combo_discount_unit"`
	ComboDiscountNums   decimal.Decimal `json:"combo_discount_nums" form:"combo_discount_nums"`
	YearDiscountUnit    int             `json:"year_discount_unit" form:"year_discount_unit"`
	YearDiscountNums    decimal.Decimal `json:"year_discount_nums" form:"year_discount_nums"`
}

type PackageScreenInfo struct {
	//Id      int64  `json:"id" form:"id"`
	PayType string `json:"pay_type" form:"pay_type"`
	Name    string `json:"name" form:"name"`
}

type MchPackageInfo struct {
	TypeName     string          `json:"type_name" form:"type_name"`
	ModelName    string          `json:"model_name" form:"model_name"`
	BusinessName string          `json:"business_name" form:"business_name"`
	ChainName    []string        `json:"chain_name" form:"chain_name"`
	SubCoinName  []string        `json:"sub_coin_name" form:"sub_coin_name"`
	Fee          []MchFeeInfo    `json:"fee" form:"fee"`
	DeductCoin   []string        `json:"deduct_coin" form:"deduct_coin"`
	DiffFee      decimal.Decimal `json:"diff_fee" form:"diff_fee"`
	TotalCost    decimal.Decimal `json:"total_cost" form:"total_cost"`
	EndCost      decimal.Decimal `json:"end_cost" form:"end_cost"`
}

type MchFeeInfo struct {
	ChainDiscountUnit int             `json:"chain_discount_unit" form:"chain_discount_unit"`
	ChainDiscountNums decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	CoinDiscountUnit  int             `json:"coin_discount_unit" form:"coin_discount_unit"`
	CoinDiscountNums  decimal.Decimal `json:"coin_discount_nums" form:"coin_discount_nums"`
	DeployFee         decimal.Decimal `json:"deploy_fee" form:"deploy_fee"`
	CoverFee          decimal.Decimal `json:"cover_fee" form:"cover_fee"`
	DiscountFee       decimal.Decimal `json:"discount_fee" form:"discount_fee"` //优惠金额
	DepositFee        decimal.Decimal `json:"deposit_fee" form:"deposit_fee"`
	AddrNums          int             `json:"addr_nums" form:"addr_nums"`
	MinerFee          string          `json:"miner_fee" form:"miner_fee"`
}
