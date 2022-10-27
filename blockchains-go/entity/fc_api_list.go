package entity

import (
	"time"
)

type FcApiList struct {
	Id    int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Title string `json:"title" xorm:"not null comment('接口名称') unique VARCHAR(64)"`
	//ApiUrl     string    `json:"api_url" xorm:"comment('接口地址，不包含主机部分') VARCHAR(100)"`
	Content    string    `json:"content" xorm:"comment('接口描述') VARCHAR(255)"`
	Status     int       `json:"status" xorm:"default 1 comment('1启用2禁用') TINYINT(4)"`
	Createtime int       `json:"createtime" xorm:"default 0 comment('创建时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	ApiSuffix  string    `json:"api_suffix" xorm:"'api_suffix'"` //api权限路径后缀
}
