package domain

import (
	"github.com/shopspring/decimal"
)

type OrderReqInfo struct {
	Id         int64  `json:"id" form:"id"`
	ContactStr string `json:"contact_str" form:"contact_str"` //联系方式
	AccountId  int64  `json:"account_id" form:"account_id"`
	BusinessId int    `json:"business_id" form:"business_id"` //业务线id
	OrderId    string `json:"order_id" form:"order_id"`       //订单流水号
	Limit      int    `json:"limit"  description:"查询条数" example:"10"`
	Offset     int    `json:"offset" description:"查询起始位置" example:"0"`
}

type OrderOperateInfo struct {
	Id      int64  `json:"id" form:"id"`
	Operate string `json:"operate" form:"operate"` //agree-同意，refuse-拒绝
	Remark  string `json:"remark" form:"remark"`   //审核备注
}

type AccountOperateInfo struct {
	Id         string `json:"id" form:"id"` //业务线订单id
	AccountId  int64  `json:"account_id" form:"account_id"`
	OrderId    string `json:"order_id" form:"order_id"`       //业务线订单id
	BusinessId int64  `json:"business_id" form:"business_id"` //业务线id
	Operate    string `json:"operate" form:"operate"`         //renew-续费 agree-同意，refuse-拒绝
	Remark     string `json:"remark" form:"remark"`           //审核备注
}

type BusinessOrderInfo struct {
	//Id                 int64           `json:"id" form:"column:id"`
	Name               string          `json:"name" form:"column:name"`
	AccountStatus      int             `json:"account_status" form:"account_status"`
	AccountId          int64           `json:"account_id" form:"account_id"`
	Email              string          `json:"email" form:"column:email"`
	Phone              string          `json:"phone" form:"column:phone"`
	OrderType          string          `json:"order_type" form:"order_type"` //交易类型
	OrderId            string          `json:"order_id" form:"order_id"`     //订单id
	TypeName           string          `json:"type_name" form:"column:type_name"`
	ModelName          string          `json:"model_name" form:"column:model_name"`
	BusinessId         int64           `json:"business_id" form:"business_id"`
	BusinessName       string          `json:"business_name" form:"business_name"`
	Coin               string          `json:"coin" form:"coin"`         //主链币
	SubCoin            string          `json:"sub_coin" form:"sub_coin"` //代币
	DeployFee          decimal.Decimal `json:"deploy_fee" form:"column:deploy_fee"`
	CustodyFee         decimal.Decimal `json:"custody_fee" form:"column:custody_fee"`
	DepositFee         decimal.Decimal `json:"deposit_fee" form:"column:deposit_fee"`
	CoverFee           decimal.Decimal `json:"cover_fee" form:"column:cover_fee"`
	AddBusinessFee     decimal.Decimal `json:"add_business_fee" form:"column:add_business_fee"`             //增加业务费
	AddChainFee        decimal.Decimal `json:"add_chain_fee" form:"column:add_chain_fee"`                   //增加主链币费
	AddSubChainFee     decimal.Decimal `json:"add_sub_chain_fee,omitempty" form:"column:add_sub_chain_fee"` //增加代币费
	DiscountFee        decimal.Decimal `json:"discount_fee" form:"column:discount_fee"`                     //优惠费
	ProfitNumber       decimal.Decimal `json:"profit_number" form:"profit_number"`                          //套餐获益户
	DeductCoin         string          `json:"deduct_coin" form:"deduct_coin"`                              //扣费币种
	DeductCoinName     string          `json:"deduct_coin_name" form:"deduct_coin_name"`                    //扣费币种
	AdminVerifyId      int64           `json:"admin_verify_id" form:"admin_verify_id"`                      //管理员审核人id
	AdminVerifyName    string          `json:"admin_verify_name" form:"admin_verify_name"`                  //管理员审核时间
	AdminVerifyTime    string          `json:"admin_verify_time" form:"admin_verify_time"`                  //管理员审核时间
	AdminVerifyState   string          `json:"admin_verify_state" form:"admin_verify_state"`                //管理员审核状态
	AccountVerifyTime  string          `json:"account_verify_time" form:"account_verify_time"`              //商户审核时间
	AccountVerifyState string          `json:"account_verify_state" form:"account_verify_state"`            //商户审核时间
	TotalFee           decimal.Decimal `json:"total_fee" form:"total_fee"`                                  //商户审核状态
	Remark             string          `json:"remark" form:"remark"`
	CreateTime         string          `json:"create_time" form:"create_time"`
}
