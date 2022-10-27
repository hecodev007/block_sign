package entity

import (
	"time"
)

type FcTransactionsEosSta struct {
	Id                  int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Grep                string    `json:"grep" xorm:"not null comment('äº¤æ˜“ç±»åž‹') VARCHAR(30)"`
	CoinName            string    `json:"coin_name" xorm:"not null comment('å¸ç§ID') VARCHAR(15)"`
	Address             string    `json:"address" xorm:"not null comment('åœ°å€') index unique(idx_unique) VARCHAR(100)"`
	Content             string    `json:"content" xorm:"TEXT"`
	DoTime              int       `json:"do_time" xorm:"not null default 0 comment('äº¤æ˜“æ—¶é—´æˆ³') INT(11)"`
	DoDate              time.Time `json:"do_date" xorm:"comment('äº¤æ˜“æ—¥æœŸ') DATE"`
	Block               int64     `json:"block" xorm:"not null default 0 comment('æ‰€åœ¨åŒºå—é«˜åº¦') BIGINT(20)"`
	Txid                string    `json:"txid" xorm:"not null comment('äº¤æ˜“ID(å”¯ä¸€é”®)') unique(idx_unique) VARCHAR(100)"`
	Type                int       `json:"type" xorm:"not null default 0 comment('åœ°å€ç±»åž‹ï¼š1å†·é’±åŒ…2çƒ­é’±åŒ…3å¼€æˆ·åœ°å€4.è®¡åˆ’å¤–åœ°å€') TINYINT(3)"`
	OppositeAddressType int       `json:"opposite_address_type" xorm:"not null default 0 comment('对方地址类型，0-未知或外部地址；1-冷钱包地址，2-热钱包地址，3-用户地址') TINYINT(2)"`
	Lastmodify          time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
