package entity

type FcConfig struct {
	Id    int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Name  string `json:"name" xorm:"unique(idx_type_name) VARCHAR(50)"`
	Value string `json:"value" xorm:"TEXT"`
	Type  string `json:"type" xorm:"unique(idx_type_name) VARCHAR(50)"`
}
