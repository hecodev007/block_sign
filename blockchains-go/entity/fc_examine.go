package entity

type FcExamine struct {
	Id              int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	EntId           int    `json:"ent_id" xorm:"INT(11)"`
	AppId           int    `json:"app_id" xorm:"not null INT(11)"`
	CoinId          string `json:"coin_id" xorm:"not null VARCHAR(20)"`
	Change          string `json:"change" xorm:"not null default 0.0000000000 comment('发生额') DECIMAL(23,10)"`
	Fee             string `json:"fee" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(23,10)"`
	Address         string `json:"address" xorm:"not null comment('发起地址') VARCHAR(255)"`
	OppositeAddress string `json:"opposite_address" xorm:"not null comment('对方地址') VARCHAR(255)"`
	Status          int    `json:"status" xorm:"not null default 0 comment('0待审核1审核成功2审核失败') TINYINT(1)"`
	Type            int    `json:"type" xorm:"comment('1.转出2.生成礼品卡') TINYINT(1)"`
	Remark          string `json:"remark" xorm:"comment('记录') VARCHAR(50)"`
	TradeTime       string `json:"trade_time" xorm:"not null comment('申请时间') DECIMAL(18,6)"`
	UpdateTime      string `json:"update_time" xorm:"not null comment('审核时间') DECIMAL(18,6)"`
	TradeId         string `json:"trade_id" xorm:"not null comment('平台交易编号') unique VARCHAR(50)"`
	Code            string `json:"code" xorm:"not null comment('范奈斯交易编号') VARCHAR(20)"`
	ChainType       int    `json:"chain_type" xorm:"not null default 0 comment('1站内，0站外') TINYINT(1)"`
	CreateTime      int    `json:"create_time" xorm:"INT(11)"`
	Total           int    `json:"total" xorm:"comment('申请礼品卡数量') INT(11)"`
	CardType        int    `json:"card_type" xorm:"comment('1个人') TINYINT(1)"`
}
