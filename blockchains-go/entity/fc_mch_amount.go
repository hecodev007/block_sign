package entity

type FcMchAmount struct {
	Id       int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinId   int    `json:"coin_id" xorm:"not null default 0 INT(10)"`
	CoinType string `json:"coin_type" xorm:"not null comment('币种名称') unique(coin_mch) VARCHAR(100)"`
	Amount   string `json:"amount" xorm:"not null default 0.000000000000000000 comment('当前余额') DECIMAL(60,24)"`
	AppId    int    `json:"app_id" xorm:"not null default 0 comment('商户id') unique(coin_mch) INT(10)"`
}
