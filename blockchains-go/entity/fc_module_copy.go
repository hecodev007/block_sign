package entity

type FcModuleCopy struct {
	ModId    int    `json:"mod_id" xorm:"not null pk autoincr SMALLINT(6)"`
	Module   string `json:"module" xorm:"default 'module' ENUM('menu','module','top')"`
	Level    int    `json:"level" xorm:"default 3 index TINYINT(1)"`
	Ctl      string `json:"ctl" xorm:"default '' VARCHAR(20)"`
	Act      string `json:"act" xorm:"default '' VARCHAR(30)"`
	Title    string `json:"title" xorm:"default '' VARCHAR(20)"`
	Visible  int    `json:"visible" xorm:"default 1 TINYINT(1)"`
	ParentId int    `json:"parent_id" xorm:"default 0 index SMALLINT(6)"`
	Orderby  int    `json:"orderby" xorm:"default 50 SMALLINT(6)"`
	Icon     string `json:"icon" xorm:"VARCHAR(30)"`
}
