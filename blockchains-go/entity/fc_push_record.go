package entity

type FcPushRecord struct {
	Id            int64  `json:"id" xorm:"pk autoincr BIGINT(15)"`
	AppId         int    `json:"app_id" xorm:"not null default 0 INT(11)"`
	Coin          string `json:"coin" xorm:"not null comment('主币种名称') VARCHAR(100)"`
	CoinType      string `json:"coin_type" xorm:"not null comment('币种名称') VARCHAR(100)"`
	Msg           string `json:"msg" xorm:"TEXT"`
	AddTime       int    `json:"add_time" xorm:"not null default 0 INT(10)"`
	Url           string `json:"url" xorm:"VARCHAR(200)"`
	TxId          string `json:"tx_id" xorm:"VARCHAR(150)"`
	Confirmations int    `json:"confirmations" xorm:"not null default 0 TINYINT(3)"`
	Status        int    `json:"status" xorm:"not null default 0 comment('1 推送成功 ') TINYINT(3)"`
}

func (f FcPushRecord) TableName() string {
	return "fc_push_record2"
}
