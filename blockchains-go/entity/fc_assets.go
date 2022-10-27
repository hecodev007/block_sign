package entity

import (
	"time"
)

type FcAssets struct {
	Id               int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Type             int       `json:"type" xorm:"not null comment('1站外转出2站外转入3平台转出4平台转入') TINYINT(1)"`
	Address          string    `json:"address" xorm:"VARCHAR(255)"`
	OppositeAddress  string    `json:"opposite_address" xorm:"VARCHAR(255)"`
	CoinId           int       `json:"coin_id" xorm:"not null INT(11)"`
	Category         int       `json:"category" xorm:"not null comment('-1站外用户0:冷钱包，其他:平台id') INT(11)"`
	OppositeCategory int       `json:"opposite_category" xorm:"not null comment('-1站外用户0:冷钱包，其他:平台id') INT(11)"`
	Before           string    `json:"before" xorm:"not null default 0.0000000000 comment('之前余额') DECIMAL(23,10)"`
	After            string    `json:"after" xorm:"not null default 0.0000000000 comment('之后余额') DECIMAL(23,10)"`
	Change           string    `json:"change" xorm:"not null default 0.0000000000 comment('改变余额') DECIMAL(23,10)"`
	Fee              string    `json:"fee" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(20,10)"`
	Addtime          time.Time `json:"addtime" xorm:"not null DATETIME"`
	Txid             string    `json:"txid" xorm:"not null VARCHAR(80)"`
	AppId            int       `json:"app_id" xorm:"INT(11)"`
	Code             string    `json:"code" xorm:"comment('交易编号') VARCHAR(20)"`
}
