package dot

import (
	"time"
)

type Notifyresult struct {
	Id        int64     `xorm:"pk autoincr BIGINT(20)"`
	Userid    int       `xorm:"not null default 0 comment('通知用户id') index(userid) INT(11)"`
	Txid      string    `xorm:"not null default '' comment('交易id') index(userid) VARCHAR(255)"`
	Num       int       `xorm:"not null default 0 comment('推送次数') INT(11)"`
	Timestamp time.Time `xorm:"'timestamp' not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	Result    int       `xorm:"not null default 0 comment('推送结果 1表示成功') INT(11)"`
	Content   string    `xorm:"not null default '' comment('失败内容') VARCHAR(1024)"`
}
