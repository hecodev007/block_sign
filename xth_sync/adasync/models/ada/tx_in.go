package ada


type TxIn struct {
	Id int64      `xorm:"pk autoincr BIGINT(20)"`
	TxInId int64
	TxOutId int64
	TxOutIndex int64
}

type TxInExtend struct{
	TxIn `xorm:"extends"`
	TxOut `xorm:"extends"`
}

func (txin *TxInExtend) TableName() string{
	return "tx_in"
}