package services

import (
	"time"
)

type ServiceEntity struct {
	Id                  int       `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Name                string    `json:"name" gorm:"column:name"`
	WithdrawalStatus    int       `json:"withdrawal_status" gorm:"column:withdrawal_status"`
	LimitSameWithdrawal int       `json:"limit_same_withdrawal" gorm:"column:limit_same_withdrawal"`
	LimitTransfer       int       `json:"limit_transfer" gorm:"column:limit_transfer"`
	AuditType           int       `json:"audit_type" gorm:"column:audit_type"`
	Phone               string    `json:"phone" gorm:"phone"`
	State               int       `json:"state" gorm:"column:state"`
	Remark              string    `json:"remark" gorm:"column:remark"`
	CreateTime          time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime          time.Time `json:"update_time" gorm:"column:update_time"`
}

func (s *ServiceEntity) TableName() string {
	return "service"
}
