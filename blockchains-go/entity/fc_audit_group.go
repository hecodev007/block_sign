package entity

import (
	"time"
)

type FcAuditGroup struct {
	LvId       int       `json:"lv_id" xorm:"not null pk autoincr INT(11)"`
	Level      int       `json:"level" xorm:"TINYINT(4)"`
	Name       string    `json:"name" xorm:"VARCHAR(50)"`
	UserId     int       `json:"user_id" xorm:"comment('审核人ID') INT(11)"`
	Addtime    int64     `json:"addtime" xorm:"BIGINT(20)"`
	Status     int       `json:"status" xorm:"default 1 TINYINT(4)"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
