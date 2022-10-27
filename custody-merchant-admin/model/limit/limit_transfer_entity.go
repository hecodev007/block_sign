package limit

import (
	"github.com/shopspring/decimal"
	"time"
)

type LimitTransfer struct {
	ServiceId    int             `gorm:"column:service_id" json:"service_id,omitempty"`
	Id           int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	CreateByUser int64           `json:"create_by_user" gorm:"column:create_by_user"`
	NumEach      decimal.Decimal `json:"num_each" gorm:"column:num_each"`
	NumDay       decimal.Decimal `json:"num_day" gorm:"column:num_day"`
	NumWeeks     decimal.Decimal `json:"num_weeks" gorm:"column:num_weeks"`
	NumMonth     decimal.Decimal `json:"num_month" gorm:"column:num_month"`
	CreateTime   time.Time       `json:"create_time" gorm:"column:create_time"`
	UpdateTime   time.Time       `json:"update_time" gorm:"column:update_time"`
}

func (l *LimitTransfer) TableName() string {
	return "limit_transfer"
}
