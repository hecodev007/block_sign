package ada

type Tx struct {
	Id int64      `xorm:"pk autoincr BIGINT(20)"`
	Hash []byte
	BlockId int64
	BlockIndex int64
	OutSum int64
	Fee int64
}
