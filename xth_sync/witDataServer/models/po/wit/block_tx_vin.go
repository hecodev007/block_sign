package wit

import (
	"time"
	"witDataServer/common/db"
	"witDataServer/common/log"
)

type BlockTxVin struct {
	Id         int64     `xorm:"not null pk autoincr INT(20)"`
	Txid       string    `xorm:"not null VARCHAR(100)"`
	VoutN      int       `xorm:"not null INT(20)"`
	Blockhash  string    `xorm:"default '' VARCHAR(64)"`
	Value      string    `xorm:"not null DECIMAL(50)"`
	Address    string    `xorm:"not null VARCHAR(100)"`
	SpendTxid  string    `xorm:"not null VARCHAR(100)"`
	Height     int       `xorm:"INT(20)"`
	Createtime time.Time `xorm:"TIMESTAMP"`
}

func TxVinRollBack(height int64) (int64, error) {
	bl := new(BlockTxVin)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTxVin(txs []*BlockTxVin) (affected int64, err error) {
	affected, err = db.SyncConn.Insert(txs)
	if err != nil {
		log.Warn(err.Error())
	}
	return affected, err
}
