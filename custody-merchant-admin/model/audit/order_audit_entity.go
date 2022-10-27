package audit

import "time"

type OrderAudit struct {
	AuditLevel  int       `json:"audit_level" gorm:"column:audit_level"`
	State       int       `json:"state" gorm:"column:state"`
	AuditResult int       `json:"audit_result" gorm:"column:audit_result"`
	Id          int64     `json:"id" gorm:"column:id; PRIMARY_KEY"`
	OrderId     int64     `json:"order_id" gorm:"column:order_id"`
	UserId      int64     `json:"user_id" gorm:"column:user_id"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time"`
}

type NumsAudit struct {
	Nums       int `json:"nums" gorm:"column:nums"`
	AuditLevel int `json:"audit_level" gorm:"column:audit_level"`
}

func (o *OrderAudit) TableName() string {
	return "order_audit"
}
