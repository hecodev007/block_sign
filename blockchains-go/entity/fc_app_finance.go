package entity

type FcAppFinance struct {
	Id              int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	EntId           int    `json:"ent_id" xorm:"not null INT(11)"`
	AppId           int    `json:"app_id" xorm:"not null index INT(11)"`
	CoinId          int    `json:"coin_id" xorm:"not null INT(11)"`
	Address         string `json:"address" xorm:"comment('发起地址') VARCHAR(255)"`
	OppositeAddress string `json:"opposite_address" xorm:"comment('对方地址') index VARCHAR(255)"`
	Type            int    `json:"type" xorm:"not null comment('1.转入2.转出3.冻结4解冻.5收益6.奖励7.生成礼品卡8.得到礼品卡9.转入增值钱包10.转出增值钱包11.退回礼品卡') unique(type) TINYINT(4)"`
	Change          string `json:"change" xorm:"default 0.0000000000 comment('操作金额') DECIMAL(23,10)"`
	Fee             string `json:"fee" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(23,10)"`
	Remark          string `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	ChainType       int    `json:"chain_type" xorm:"not null default 0 comment('1站内，0站外') TINYINT(4)"`
	Txid            string `json:"txid" xorm:"VARCHAR(80)"`
	Confirm         int    `json:"confirm" xorm:"default 0 comment('确认数') INT(11)"`
	Code            string `json:"code" xorm:"comment('范奈斯交易编号') VARCHAR(50)"`
	TradeId         string `json:"trade_id" xorm:"not null comment('平台交易编号') unique(type) VARCHAR(255)"`
	TradeTime       string `json:"trade_time" xorm:"not null comment('申请时间') DECIMAL(18,6)"`
	Dealed          int    `json:"dealed" xorm:"default 0 comment('是否生成过快照：0否，1是') TINYINT(3)"`
	CreateTime      int    `json:"create_time" xorm:"INT(11)"`
	IsExamice       int    `json:"is_examice" xorm:"not null default 0 comment('0不审核1审核') TINYINT(1)"`
}
