package entity

type FcApiRules struct {
	ModId   int    `json:"mod_id" xorm:"not null pk autoincr SMALLINT(6)"`
	Level   int    `json:"level" xorm:"default 3 index TINYINT(1)"`
	Ctl     string `json:"ctl" xorm:"default '' VARCHAR(20)"`
	Act     string `json:"act" xorm:"default '' VARCHAR(30)"`
	Title   string `json:"title" xorm:"default '' VARCHAR(20)"`
	Orderby int    `json:"orderby" xorm:"default 50 SMALLINT(6)"`
}
