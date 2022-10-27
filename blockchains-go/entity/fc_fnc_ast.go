package entity

type FcFncAst struct {
	Id              int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address         string `json:"address" xorm:"not null comment('发起地址') VARCHAR(255)"`
	OppositeAddress string `json:"opposite_address" xorm:"not null comment('对方地址') index VARCHAR(255)"`
	TypeTxt         string `json:"type_txt" xorm:"not null default '0' VARCHAR(50)"`
	Coin            string `json:"coin" xorm:"not null default '0' VARCHAR(20)"`
	Amount          string `json:"amount" xorm:"not null default 0.0000000000 comment('操作金额') DECIMAL(23,10)"`
	Fee             string `json:"fee" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(23,10)"`
	TradeTime       string `json:"trade_time" xorm:"comment('交易时间') VARCHAR(20)"`
}
