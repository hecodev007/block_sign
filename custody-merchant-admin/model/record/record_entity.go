package record

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

const (
	FinanceRecord  = "finance"  //财务
	BusinessRecord = "business" //业务线
	ApplyRecord    = "apply"    //申请
)

//Entity 操作日志表
type Entity struct {
	Db           *orm.CacheDB `json:"-" gorm:"-"`
	Id           int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	OperatorId   int64        `json:"operator_id" gorm:"column:operator_id"`
	OperatorName string       `json:"operator_name" gorm:"column:operator_name"`
	Operate      string       `json:"operate" gorm:"column:operate"` //当财务日志 时 冻结用户和资产lock_user，解冻用户和资产unlock_user，冻结资产lock_asset，，解冻资产unlock_asset
	Remark       string       `json:"remark" gorm:"column:remark"`
	BusinessId   int64        `json:"business_id" gorm:"column:business_id"`
	MerchantId   int64        `json:"merchant_id" gorm:"column:merchant_id"`
	FinanceId    int64        `json:"finance_id" gorm:"column:finance_id"`
	RecordType   string       `json:"record_type" gorm:"column:record_type"` //操作类型finance-财务管理操作，business-操作
	CreatedAt    time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_record"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
