package entity

import (
	"time"
)

type FcAdminLoginLog struct {
	Id        int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Ip        string    `json:"ip" xorm:"not null index VARCHAR(20)"`
	Addtime   time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	Username  string    `json:"username" xorm:"index VARCHAR(20)"`
	GroupId   int       `json:"group_id" xorm:"not null INT(11)"`
	DeviceSn  string    `json:"device_sn" xorm:"not null default '' VARCHAR(64)"`
	Lng       float64   `json:"lng" xorm:"default 0.00000000 DOUBLE(12,8)"`
	Lat       float64   `json:"lat" xorm:"default 0.00000000 DOUBLE(12,8)"`
	Address   string    `json:"address" xorm:"VARCHAR(255)"`
	Ua        string    `json:"ua" xorm:"default 'browser' VARCHAR(200)"`
	FromCli   string    `json:"from_cli" xorm:"not null comment('客户端来源：ios, browser') VARCHAR(50)"`
	LoginFace string    `json:"login_face" xorm:"comment('登录者人脸图像') VARCHAR(200)"`
}
