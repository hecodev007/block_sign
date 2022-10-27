package avax

import (
	"time"
)

type Notifyresult struct {
	Id        int64     `xorm:"pk autoincr BIGINT(20)"`
	Userid    int       `xorm:"not null default 0 comment('é€šçŸ¥ç”¨æˆ·id') index(userid) INT(11)"`
	Height    int64     `xorm:"not null comment('é«˜åº¦') BIGINT(20)"`
	Txid      string    `xorm:"not null default '' comment('äº¤æ˜“id') index(userid) VARCHAR(255)"`
	Num       int       `xorm:"not null default 0 comment('æŽ¨é€æ¬¡æ•°') INT(11)"`
	Timestamp time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('æŽ¨é€æ—¶é—´') TIMESTAMP"`
	Result    int       `xorm:"not null default 0 comment('æŽ¨é€ç»“æžœ 1è¡¨ç¤ºæˆåŠŸ') INT(11)"`
	Content   string    `xorm:"not null default '' comment('å¤±è´¥å†…å®¹') VARCHAR(1024)"`
	Type      int       `xorm:"comment('0ä¸ºæ™®é€šäº¤æ˜“æŽ¨é€ï¼Œï¼‘ä¸ºç¡®è®¤æ•°æŽ¨é€') TINYINT(3)"`
}
