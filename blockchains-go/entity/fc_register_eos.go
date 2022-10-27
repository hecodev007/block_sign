package entity

type FcRegisterEos struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(10)"`
	Username   string `json:"username" xorm:"comment('用户名') index VARCHAR(20)"`
	Amount     string `json:"amount" xorm:"default 0.0000000000 comment('金额') DECIMAL(23,10)"`
	Status     int    `json:"status" xorm:"comment('1成功0失败') TINYINT(4)"`
	TradeId    string `json:"trade_id" xorm:"comment('交易编号') unique VARCHAR(100)"`
	AddTime    int    `json:"add_time" xorm:"INT(11)"`
	FatherName string `json:"father_name" xorm:"comment('父级用户名') index VARCHAR(20)"`
	Content    string `json:"content" xorm:"not null comment('备注') TEXT"`
}
