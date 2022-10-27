package ada

type TxOut struct {
	Id int64      `xorm:"pk autoincr BIGINT(20)"`
	TxId int64
	Index int64
	Address string
	Value int64
}
