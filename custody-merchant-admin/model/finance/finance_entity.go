package finance

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

//Entity 商户财务审核列表
type Entity struct {
	Db            *orm.CacheDB `json:"-" gorm:"-"`
	Id            int64        `json:"id" gorm:"id"`
	AccountId     int64        `json:"account_id" gorm:"account_id"`
	ApplyId       int64        `json:"apply_id" gorm:"apply_id"`           //申请表-合同信息等
	VerifyStatus  string       `json:"verify_status" gorm:"verify_status"` //财务审核状态 空-未审核，agree-已审核通过，refuse-已拒绝
	VerifyAt      *time.Time   `json:"verify_at" gorm:"verify_at"`
	VerifyUser    string       `json:"verify_user" gorm:"verify_user"`
	IsLock        int          `json:"is_lock" gorm:"is_lock"`                 //是否冻结 1-冻结，0-未冻结
	IsLockFinance int          `json:"is_lock_finance" gorm:"is_lock_finance"` //是否冻结资产 1-冻结，0-未冻结
	Remark        string       `json:"remark" gorm:"remark"`
	CreatedAt     time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"updated_at"`
}

func (e *Entity) TableName() string {
	return "service_finance"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type FinanceListDB struct {
	Id              int64     `json:"id" gorm:"id"`
	AccountId       int64     `json:"account_id" gorm:"account_id"`
	AccountName     string    `json:"account_name" gorm:"account_name"`
	Phone           string    `json:"phone" gorm:"phone"`
	Email           string    `json:"email" gorm:"email"`
	IdCardNum       string    `json:"id_card_num" gorm:"id_card_num"`
	PassportNum     string    `json:"passport_num" gorm:"passport_num"`
	CreatedAt       time.Time `json:"created_at" gorm:"created_at"`
	RealNameStatus  string    `json:"real_name_status" gorm:"real_name_status"`
	RealNameAt      time.Time `json:"real_name_at" gorm:"real_name_at"`
	ContractStartAt time.Time `json:"contract_start_at" gorm:"contract_start_at"`
	ContractEndAt   time.Time `json:"contract_end_at" gorm:"contract_end_at"`
	FvStatus        string    `json:"fv_status" gorm:"fv_status"`           //财务审核状态
	FVRemark        string    `json:"fv_remark" gorm:"fv_remark"`           //财务操作备注
	LockStatus      string    `json:"lock_status" gorm:"lock_status"`       //账户冻结状态
	LockRemark      string    `json:"lock_remark" gorm:"lock_remark"`       //账户冻结备注
	AccountStatus   int       `json:"account_status" gorm:"account_status"` //用户状态（0-正式，1-测试）
	TestEnd         string    `json:"test_end" gorm:"test_end"`
	TradeType       string    `json:"trade_type" gorm:"column:trade_type"`    //交易类型(open-开通，2renew-二次续费，3renew-三次续费。。。。)
	IsLockFinance   int       `json:"is_lock_finance" gorm:"is_lock_finance"` //用户状态（0-正式，1-测试）
	IsLock          int       `json:"is_lock" gorm:"is_lock"`                 //用户状态（0-正式，1-测试）
}
