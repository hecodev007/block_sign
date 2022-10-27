package dot

type BlockAmount struct {
	Id      int64  `xorm:"pk autoincr BIGINT(20)"`
	Address string `xorm:"not null default '' comment('账号地址') index VARCHAR(100)"`
	Amount  int64  `xorm:"not null default 0 comment('金额') BIGINT(40)"`
}
