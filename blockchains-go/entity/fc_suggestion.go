package entity

import (
	"time"
)

type FcSuggestion struct {
	Id       int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Username string    `json:"username" xorm:"not null comment('用户名') index VARCHAR(25)"`
	Title    string    `json:"title" xorm:"comment('标题') VARCHAR(200)"`
	Content  string    `json:"content" xorm:"comment('内容') VARCHAR(500)"`
	Images   string    `json:"images" xorm:"comment('图片') VARCHAR(1000)"`
	Addtime  time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP comment('时间') TIMESTAMP"`
	Status   int       `json:"status" xorm:"default 0 comment('状态') TINYINT(4)"`
}
