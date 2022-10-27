package entity

import (
	"time"
)

type FcColdVerHuobiOrig struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	Content    string    `json:"content" xorm:"comment('原始数据') TEXT"`
}
