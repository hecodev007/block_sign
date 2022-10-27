package entity

import (
	"time"
)

type FcCallback struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Moudle     string    `json:"moudle" xorm:"VARCHAR(255)"`
	Action     string    `json:"action" xorm:"VARCHAR(255)"`
	Status     int       `json:"status" xorm:"not null default 0 comment('1已完成0未完成') TINYINT(4)"`
	Num        int       `json:"num" xorm:"not null default 0 comment('请求次数') INT(11)"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	Addtime    time.Time `json:"addtime" xorm:"DATETIME"`
	Msg        string    `json:"msg" xorm:"comment('返回结果') VARCHAR(255)"`
	Content    string    `json:"content" xorm:"comment('传递内容') VARCHAR(255)"`
	Pid        int       `json:"pid" xorm:"INT(11)"`
	Url        string    `json:"url" xorm:"VARCHAR(255)"`
}
