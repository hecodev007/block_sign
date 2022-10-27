package ada

type Block struct {
	Id int64      `xorm:"pk autoincr BIGINT(20)"`
	Hash []byte
	BlockNo int64
	TxCount int64
	EpochNo int64
	SlotNo int64
}
