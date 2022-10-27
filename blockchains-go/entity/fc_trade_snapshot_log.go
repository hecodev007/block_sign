package entity

import (
	"time"
)

type FcTradeSnapshotLog struct {
	Id       int64     `json:"id" xorm:"pk autoincr BIGINT(20)"`
	Expected int       `json:"expected" xorm:"default 0 comment('预期处理数量') INT(10)"`
	Actual   int       `json:"actual" xorm:"default 0 comment('实际处理的数量') INT(10)"`
	Sql      string    `json:"sql" xorm:"TEXT"`
	Date     time.Time `json:"date" xorm:"comment('日期') DATE"`
	Time     time.Time `json:"time" xorm:"not null default CURRENT_TIMESTAMP comment('时间') TIMESTAMP"`
}
