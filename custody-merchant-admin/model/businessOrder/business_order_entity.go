package businessOrder

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 业务线订单表
type Entity struct {
	Db                 *orm.CacheDB    `json:"-" gorm:"-"`
	Id                 int64           `json:"id" gorm:"id"`
	AccountId          int64           `json:"account_id" gorm:"account_id"`
	OrderType          string          `json:"order_type" gorm:"order_type"`                                //交易类型
	OrderId            string          `json:"order_id" gorm:"order_id"`                                    //订单id
	BusinessId         int64           `json:"business_id" gorm:"business_id"`                              //业务线id
	PackageId          int64           `json:"package_id" gorm:"package_id"`                                //套餐id
	BusinessPackageId  int64           `json:"business_package_id" gorm:"business_package_id"`              //业务线的套餐id
	AddBusinessFee     decimal.Decimal `json:"add_business_fee,omitempty" gorm:"column:add_business_fee"`   //增加业务费
	AddChainFee        decimal.Decimal `json:"add_chain_fee,omitempty" gorm:"column:add_chain_fee"`         //增加主链币费
	AddSubChainFee     decimal.Decimal `json:"add_sub_chain_fee,omitempty" gorm:"column:add_sub_chain_fee"` //增加代币费
	FullYearFee        decimal.Decimal `json:"full_year_fee,omitempty" gorm:"column:full_year_fee"`         //满年优惠费.判断支付订单时是否依旧满足满年优惠
	DiscountFee        decimal.Decimal `json:"discount_fee,omitempty" gorm:"column:discount_fee"`           //优惠费
	OrderCoinId        int64           `json:"order_coin_id,omitempty" gorm:"column:order_coin_id"`         //订单币种id
	OrderCoinName      string          `json:"order_coin_name,omitempty" gorm:"column:order_coin_name"`     //订单币种
	ProfitNumber       decimal.Decimal `json:"profit_number" gorm:"profit_number"`                          //套餐获益户
	AdminVerifyId      int64           `json:"admin_verify_id" gorm:"admin_verify_id"`                      //管理员审核人id
	AdminVerifyTime    *time.Time      `json:"admin_verify_time" gorm:"admin_verify_time"`                  //管理员审核时间
	AdminVerifyName    string          `json:"admin_verify_name" gorm:"admin_verify_name"`                  //管理员审核状态
	AdminVerifyState   string          `json:"admin_verify_state" gorm:"admin_verify_state"`                //管理员审核状态
	DeductState        string          `json:"deduct_state" gorm:"deduct_state"`                            //钱包扣款状态 wallet-已提交至钱包，success-扣款成功
	AccountVerifyTime  *time.Time      `json:"account_verify_time" gorm:"account_verify_time"`              //商户审核时间
	AccountVerifyState string          `json:"account_verify_state" gorm:"account_verify_state"`            //商户审核状态
	AdminRemark        string          `json:"admin_remark" gorm:"admin_remark"`
	AccountRemark      string          `json:"account_remark" gorm:"account_remark"`
	CreatedAt          time.Time       `json:"created_at" gorm:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"updated_at"`
	DeletedAt          *time.Time      `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_service_order"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type Item struct {
	Id                 int64           `json:"id" gorm:"column:id"`
	Name               string          `json:"name" gorm:"column:name"`
	AccountStatus      int             `json:"account_status" gorm:"account_status"`
	AccountId          int64           `json:"account_id" gorm:"account_id"`
	Email              string          `json:"email" gorm:"column:email"`
	Phone              string          `json:"phone" gorm:"column:phone"`
	OrderType          string          `json:"order_type" gorm:"order_type"` //交易类型
	OrderId            string          `json:"order_id" gorm:"order_id"`     //订单id
	TypeName           string          `json:"type_name,omitempty" gorm:"column:type_name"`
	ModelName          string          `json:"model_name,omitempty" gorm:"column:model_name"`
	BusinessId         int64           `json:"business_id" gorm:"business_id"`
	BusinessName       string          `json:"business_name" gorm:"business_name"`
	Coin               string          `json:"coin" gorm:"coin"`                              //主链币
	SubCoin            string          `json:"sub_coin" gorm:"sub_coin"`                      //代币
	OrderCoinName      string          `json:"order_coin_name" gorm:"column:order_coin_name"` //订单币种
	DeployFee          decimal.Decimal `json:"deploy_fee,omitempty" gorm:"column:deploy_fee"`
	CustodyFee         decimal.Decimal `json:"custody_fee,omitempty" gorm:"column:custody_fee"`
	DepositFee         decimal.Decimal `json:"deposit_fee,omitempty" gorm:"column:deposit_fee"`
	CoverFee           decimal.Decimal `json:"cover_fee,omitempty" gorm:"column:cover_fee"`
	AddBusinessFee     decimal.Decimal `json:"add_business_fee,omitempty" gorm:"column:add_business_fee"`   //增加业务费
	AddChainFee        decimal.Decimal `json:"add_chain_fee,omitempty" gorm:"column:add_chain_fee"`         //增加主链币费
	AddSubChainFee     decimal.Decimal `json:"add_sub_chain_fee,omitempty" gorm:"column:add_sub_chain_fee"` //增加代币费
	DiscountFee        decimal.Decimal `json:"discount_fee,omitempty" gorm:"column:discount_fee"`           //优惠费
	ProfitNumber       decimal.Decimal `json:"profit_number" gorm:"profit_number"`                          //套餐获益户
	DeductCoin         string          `json:"deduct_coin" form:"deduct_coin"`                              //扣费币种
	AdminVerifyId      int64           `json:"admin_verify_id" gorm:"admin_verify_id"`                      //管理员审核人id
	AdminVerifyTime    time.Time       `json:"admin_verify_time" gorm:"admin_verify_time"`                  //管理员审核时间
	AdminVerifyState   string          `json:"admin_verify_state" gorm:"admin_verify_state"`                //管理员审核状态
	AccountVerifyTime  time.Time       `json:"account_verify_time" gorm:"account_verify_time"`              //商户审核时间
	AccountVerifyState string          `json:"account_verify_state" gorm:"account_verify_state"`            //商户审核状态
	Remark             string          `json:"remark" gorm:"remark"`
	CreateTime         time.Time       `json:"create_time" gorm:"create_time"`
	CreatedAt          time.Time       `json:"created_at" gorm:"created_at"`
}
