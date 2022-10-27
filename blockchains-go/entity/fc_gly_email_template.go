package entity

type FcGlyEmailTemplate struct {
	Type    string `json:"type" xorm:"VARCHAR(20)"`
	Status  int    `json:"status" xorm:"TINYINT(1)"`
	Content string `json:"content" xorm:"TEXT"`
	Back    string `json:"back" xorm:"comment('说明') VARCHAR(100)"`
}
