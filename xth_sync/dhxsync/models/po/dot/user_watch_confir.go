package dot

type UserWatchConfir struct {
	Id     int64 `xorm:"pk autoincr BIGINT(20)"`
	Height int64 `xorm:"not null default 0 index BIGINT(20)"`
	Userid int64 `xorm:"not null default 0 BIGINT(20)"`
}
