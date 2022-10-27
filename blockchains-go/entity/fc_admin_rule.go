package entity

type FcAdminRule struct {
	Id       int    `json:"id" xorm:"not null pk autoincr INT(10)"`
	Title    string `json:"title" xorm:"VARCHAR(50)"`
	Path     string `json:"path" xorm:"VARCHAR(100)"`
	Flow     string `json:"flow" xorm:"VARCHAR(50)"`
	Sign     string `json:"sign" xorm:"VARCHAR(50)"`
	ParentId int    `json:"parent_id" xorm:"not null default 0 INT(10)"`
	Sort     int    `json:"sort" xorm:"not null default 0 comment('排序') INT(10)"`
}
