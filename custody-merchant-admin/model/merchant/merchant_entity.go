package merchant

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

//Entity 商户信息
type Entity struct {
	Db           *orm.CacheDB `json:"-" gorm:"-"`
	Sex          int          `json:"sex" gorm:"column:sex"`
	State        int          `json:"state" gorm:"column:state"` //0-正常，1-冻结，2-失效
	Id           int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Pid          int64        `json:"pid" gorm:"column:pid"`
	ApplyId      int64        `json:"apply_id" gorm:"column:apply_id"`
	Name         string       `json:"name" gorm:"column:name"`
	Email        string       `json:"email" gorm:"column:email"`
	Phone        string       `json:"phone" gorm:"column:phone"`
	PhoneCode    string       `json:"phone_code" gorm:"column:phone_code"`
	Password     string       `json:"password" gorm:"column:password"`
	Salt         string       `json:"salt" gorm:"column:salt"`
	Role         int          `json:"roles" gorm:"column:roles"`
	Remark       string       `json:"remark" gorm:"column:remark"`
	Reason       string       `json:"reason" gorm:"column:reason"`
	Passport     string       `json:"passport" gorm:"column:passport"`
	Identity     string       `json:"identity" gorm:"column:identity"`
	PwdErr       int          `json:"pwd_err" gorm:"column:pwd_err"`
	IsTest       int          `json:"is_test" gorm:"column:is_test"`
	TestTime     *time.Time   `json:"test_time" gorm:"column:test_time"`
	IsPush       int          `json:"is_push" gorm:"is_push"` //0-未推送 1-推送中 2-已推送
	PhoneCodeErr int          `json:"phone_code_err" gorm:"column:phone_code_err"`
	EmailCodeErr int          `json:"email_code_err" gorm:"column:email_code_err"`
	LoginTime    *time.Time   `json:"login_time" gorm:"column:login_time"`
	CreateTime   *time.Time   `json:"create_time" gorm:"column:create_time"`
	UpdateTime   *time.Time   `json:"update_time" gorm:"column:update_time"`
}

func (e *Entity) TableName() string {
	return "user_info"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

//MerchantApply 商户信息申请信息列表
type MerchantApply struct {
	Db              *orm.CacheDB `json:"-" gorm:"-"`
	Id              int64        `json:"id" gorm:"id"`
	ApplyId         int64        `json:"apply_id" gorm:"apply_id"`
	AccountId       int64        `json:"account_id" gorm:"account_id"`
	AccountName     string       `json:"account_name" gorm:"account_name"`
	Name            string       `json:"name" gorm:"name"`
	Phone           string       `json:"phone" gorm:"phone"`
	Email           string       `json:"email" gorm:"email"`
	CoinName        string       `json:"coin_name" gorm:"coin_name"`
	VerifyStatus    string       `json:"verify_status" gorm:"verify_status"` // 空-未审核，agree-已审核通过，refuse-已拒绝
	VerifyAt        time.Time    `json:"verify_at" gorm:"verify_at"`
	VerifyUser      string       `json:"verify_user" gorm:"verify_user"`
	TestEnd         string       `json:"test_end" gorm:"test_end"`
	ContractStartAt time.Time    `json:"contract_start_at" gorm:"contract_start_at"`
	ContractEndAt   time.Time    `json:"contract_end_at" gorm:"contract_end_at"`
	FvStatus        string       `json:"fv_status" gorm:"fv_status"`               //财务审核状态
	Introduce       string       `json:"introduce" gorm:"introduce"`               //公司介绍
	IdCardPicture   string       `json:"id_card_picture" gorm:"id_card_picture"`   //身份证件
	BusinessPicture string       `json:"business_picture" gorm:"business_picture"` //营业执照证
	ContractPicture string       `json:"contract_picture" gorm:"contract_picture"` //合同图片
	Remark          string       `json:"remark" gorm:"remark"`
	CreatedAt       time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" gorm:"updated_at"`
	IsTest          int          `json:"is_test" gorm:"column:is_test"`
	IsPush          int          `json:"is_push" gorm:"is_push"`
	Sex             int          `json:"sex" form:"sex"`
	IdCardNum       string       `json:"id_card_num" gorm:"id_card_num"`
	PassportNum     string       `json:"passport_num" gorm:"passport_num"`
	CardType        string       `json:"card_type" gorm:"card_type"`
	AccountStatus   string       `json:"account_status" gorm:"account_status"`
	RealNameStatus  int          `json:"real_name_status" gorm:"real_name_status"` //实名状态 0-未实名，1-已实名
	RealNameAt      time.Time    `json:"real_name_at" gorm:"real_name_at"`
	LockStatus      int          `json:"lock_status" gorm:"lock_status"` //冻结状态 0-未冻结，1-已冻结

}
