package services

import (
	"github.com/shopspring/decimal"
)

type ServiceAuditConfig struct {
	Id          int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	ServiceId   int             `gorm:"column:service_id" json:"service_id,omitempty"`
	ServiceName string          `gorm:"column:service_name" json:"service_name,omitempty"`
	AuditLevel  int             `gorm:"column:audit_level" json:"audit_level,omitempty"`
	AuditType   int             `gorm:"column:audit_type" json:"audit_type,omitempty"`
	Users       string          `gorm:"column:users" json:"users,omitempty"`
	LimitUse    int             `gorm:"column:limit_use" json:"limit_use,omitempty"`
	NumEach     decimal.Decimal `json:"num_each" gorm:"column:num_each"`
	NumDay      decimal.Decimal `json:"num_day" gorm:"column:num_day"`
	NumWeek     decimal.Decimal `json:"num_week" gorm:"column:num_week"`
	NumMonth    decimal.Decimal `json:"num_month" gorm:"column:num_month"`
	State       int             `gorm:"column:state" json:"state,omitempty"`
	Reason      string          `gorm:"column:reason" json:"reason,omitempty"`
}

func (u *ServiceAuditConfig) TableName() string {
	return "service_audit_config"
}
