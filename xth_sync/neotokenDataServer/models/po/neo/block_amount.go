package neo

type BlockAmount struct {
	Id      int64  `xorm:"pk autoincr BIGINT(20)"`
	Address string `xorm:"not null default '' comment('账号地址') index VARCHAR(100)"`
	Amount  string `xorm:"not null default 0.000000000000000000 comment('金额') DECIMAL(40,18)"`
}
