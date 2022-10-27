package business

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 业务线表
type Entity struct {
	Db                  *orm.CacheDB `json:"-" gorm:"-"`
	Id                  int64        `json:"id" gorm:"id"`
	Name                string       `json:"name" gorm:"name"` //业务线名称
	Phone               string       `json:"phone" gorm:"phone"`
	Email               string       `json:"email" gorm:"email"`
	AccountId           int64        `json:"account_id" gorm:"account_id"`
	AccountStatus       int          `json:"account_status" gorm:"account_status"` //创建业务线时的用户状态
	FounderId           int64        `json:"founder_id" gorm:"founder_id"`
	CheckerId           int64        `json:"checker_id" gorm:"checker_id"`
	CheckerName         string       `json:"checker_name" gorm:"checker_name"`
	Remark              string       `json:"remark" gorm:"remark"`
	Coin                string       `json:"coin" gorm:"coin"`         //主链币
	SubCoin             string       `json:"sub_coin" gorm:"sub_coin"` //代币
	CheckedAt           *time.Time   `json:"checked_at" gorm:"checked_at"`
	WithdrawalStatus    int          `json:"withdrawal_status" gorm:"withdrawal_status"`
	LimitSameWithdrawal int          `json:"limit_same_withdrawal" gorm:"column:limit_same_withdrawal"`
	LimitTransfer       int          `json:"limit_transfer" gorm:"limit_transfer"`
	AuditType           int          `json:"audit_type" gorm:"audit_type"`
	State               int          `json:"state" gorm:"state"` //'状态：0/有效，1/冻结，2/无效'
	CreateTime          *time.Time   `json:"create_time" gorm:"create_time"`
	UpdateTime          *time.Time   `json:"update_time" gorm:"update_time"`
}

func (e *Entity) TableName() string {
	return "service"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type BPDetailInfo struct {
	Db                *orm.CacheDB    `json:"-" gorm:"-"`
	Id                int64           `json:"id" gorm:"id"`
	Phone             string          `json:"phone" gorm:"phone"`
	Email             string          `json:"email" gorm:"email"`
	Coin              string          `json:"coin" gorm:"coin"`
	SubCoin           string          `json:"sub_coin" gorm:"sub_coin"`
	AccountId         int64           `json:"account_id" gorm:"account_id"`
	AccountStatus     int             `json:"account_status" gorm:"account_status"`
	BusinessStatus    int             `json:"business_status" gorm:"business_status"`
	BusinessName      string          `json:"business_name" gorm:"business_name"`
	DeductCoin        string          `json:"deduct_coin,omitempty" gorm:"column:deduct_coin"` //扣费币种
	FounderId         int64           `json:"founder_id" gorm:"founder_id"`
	CheckerId         int64           `json:"checker_id" gorm:"checker_id"`
	CheckerName       string          `json:"checker_name" gorm:"checker_name"`
	Remark            string          `json:"remark" gorm:"remark"`
	CheckedAt         *time.Time      `json:"checked_at" gorm:"checked_at"`
	WithdrawalStatus  int             `json:"withdrawal_status" gorm:"withdrawal_status"`
	LimitTransfer     int             `json:"limit_transfer" gorm:"limit_transfer"`
	AuditType         int             `json:"audit_type" gorm:"audit_type"`
	State             int             `json:"state" gorm:"state"`
	BusinessId        int64           `json:"business_id" gorm:"column:business_id"`
	TypeName          string          `json:"type_name" gorm:"column:type_name"`
	ModelName         string          `json:"model_name" gorm:"column:model_name"`
	EnterUnit         int             `json:"enter_unit" gorm:"column:enter_unit"`
	LimitType         int             `json:"limit_type" gorm:"column:limit_type"`
	TypeNums          decimal.Decimal `json:"type_nums" gorm:"column:type_nums"`
	TopUpType         int             `json:"top_up_type" gorm:"column:top_up_type"`
	TopUpFee          string          `json:"top_up_fee" gorm:"column:top_up_fee"`
	WithdrawalType    int             `json:"withdrawal_type" gorm:"column:withdrawal_type"`
	WithdrawalFee     string          `json:"withdrawal_fee" gorm:"column:withdrawal_fee"`
	ChainNums         int             `json:"chain_nums" gorm:"column:chain_nums"`
	ChainDiscountUnit int             `json:"chain_discount_unit" gorm:"column:chain_discount_unit"`
	ChainDiscountNums decimal.Decimal `json:"chain_discount_nums" gorm:"column:chain_discount_nums"`
	ChainTimeUnit     int             `json:"chain_time_unit" gorm:"column:chain_time_unit"`
	CoinNums          int             `json:"coin_nums" gorm:"column:coin_nums"`
	CoinDiscountUnit  int             `json:"coin_discount_unit" gorm:"column:coin_discount_unit"`
	CoinDiscountNums  decimal.Decimal `json:"coin_discount_nums" gorm:"column:coin_discount_nums"`
	CoinTimeUnit      int             `json:"coin_time_unit" gorm:"column:coin_time_unit"`
	DeployFee         decimal.Decimal `json:"deploy_fee" gorm:"column:deploy_fee"`
	CustodyFee        decimal.Decimal `json:"custody_fee" gorm:"column:custody_fee"`
	DepositFee        decimal.Decimal `json:"deposit_fee" gorm:"column:deposit_fee"`
	AddrNums          int             `json:"addr_nums" gorm:"column:addr_nums"`
	CoverFee          decimal.Decimal `json:"cover_fee" gorm:"column:cover_fee"`
	ComboDiscountUnit int             `json:"combo_discount_unit" gorm:"column:combo_discount_unit"`
	ComboDiscountNums decimal.Decimal `json:"combo_discount_nums" gorm:"column:combo_discount_nums"`
	YearDiscountUnit  int             `json:"year_discount_unit" gorm:"column:year_discount_unit"`
	YearDiscountNums  decimal.Decimal `json:"year_discount_nums" gorm:"column:year_discount_nums"`
	CreatedAt         time.Time       `json:"created_at" gorm:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"updated_at"`
	DeletedAt         *time.Time      `json:"deleted_at" gorm:"deleted_at"`
	ClientId          string          `json:"client_id" gorm:"client_id"`
	Secret            string          `json:"secret" gorm:"secret"`
	IpAddr            string          `json:"ip_addr" gorm:"ip_addr"`
	CallbackUrl       string          `json:"callback_url" gorm:"callback_url"`
	IsSms             int             `json:"is_sms" gorm:"is_sms"`
	IsEmail           int             `json:"is_email" gorm:"is_email"`
	IsWithdrawal      int             `json:"is_withdrawal" gorm:"is_withdrawal"`
	IsWhitelist       int             `json:"is_whitelist" gorm:"is_whitelist"`
	IsIp              int             `json:"is_ip" gorm:"is_ip"`
	IsPlatformCheck   int             `json:"is_platform_check" gorm:"is_platform_check"`
	IsAccountCheck    int             `json:"is_account_check" gorm:"is_account_check"`
}

type BusinessListDB struct {
	AccountId      int64           `json:"account_id" form:"account_id"`
	Name           string          `json:"name" gorm:"column:name"`
	Email          string          `json:"email" gorm:"column:email"`
	Phone          string          `json:"phone" gorm:"column:phone"`
	AccountStatus  int             `json:"account_status" form:"account_status"`
	BusinessName   string          `json:"business_name" form:"business_name"`
	BusinessId     int             `json:"business_id" form:"business_id"`
	CreateTime     time.Time       `json:"create_time" gorm:"create_time"`
	Coin           string          `json:"coin" gorm:"coin"`                            //主链币
	SubCoin        string          `json:"sub_coin" gorm:"sub_coin"`                    //代币
	TypeName       string          `json:"type_name,omitempty" gorm:"column:type_name"` //套餐类型
	ModelName      string          `json:"model_name" gorm:"column:model_name"`
	ProfitNumber   decimal.Decimal `json:"profit_number" gorm:"profit_number"` //套餐获益户
	OrderType      string          `json:"order_type" gorm:"order_type"`       //交易类型
	TopUpType      int             `json:"top_up_type,omitempty" gorm:"column:top_up_type"`
	TopUpFee       string          `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"`
	WithdrawalType int             `json:"withdrawal_type,omitempty" gorm:"column:withdrawal_type"`
	WithdrawalFee  string          `json:"withdrawal_fee,omitempty" gorm:"column:withdrawal_fee"`
	CheckerName    string          `json:"checker_name" gorm:"checker_name"`
	BusinessStatus int             `json:"business_status" form:"business_status"`
	CheckedAt      time.Time       `json:"checked_at" gorm:"checked_at"`
	Remark         string          `json:"remark" form:"remark"` //审核备注

}
