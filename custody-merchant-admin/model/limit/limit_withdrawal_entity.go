package limit

import (
	"github.com/shopspring/decimal"
	"time"
)

type LimitWithdrawal struct {
	ServiceId    int             `json:"service_id,omitempty" gorm:"column:service_id" `
	NumMinutes   int             `json:"num_minutes" gorm:"column:num_minutes"`
	NumHours     int             `json:"num_hours" gorm:"column:num_hours"`
	Id           int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	CreateByUser int64           `json:"create_by_user" gorm:"column:create_by_user"`
	LineMinutes  decimal.Decimal `json:"line_minutes" gorm:"column:line_minutes"`
	LineHours    decimal.Decimal `json:"line_hours" gorm:"column:line_hours"`
	CreateTime   time.Time       `json:"create_time" gorm:"column:create_time"`
	UpdateTime   time.Time       `json:"update_time" gorm:"column:update_time"`
}

func (l *LimitWithdrawal) TableName() string {
	return "limit_withdrawal"
}
