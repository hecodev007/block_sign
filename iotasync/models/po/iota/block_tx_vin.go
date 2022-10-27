package iota

import (
	"iotasync/common/db"
	"iotasync/common/log"
	"time"
)

type BlockTxVin struct {
	Id          int64     `xorm:"not null pk autoincr INT(20)"`
	MessageId   string    `xorm:"not null VARCHAR(100)"`
	Txid        string    `xorm:"not null VARCHAR(100)"`
	OutputIndex uint16    `xorm:"INT(20)"`
	LedgerIndex uint64    `xorm:"INT(20)"`
	Spent       bool      `xorm:"tinyint(4)"`
	Value       string    `xorm:"not null DECIMAL(50)"`
	Address     string    `xorm:"not null VARCHAR(100)"`
	Height      int64     `xorm:"INT(20)"`
	Createtime  time.Time `xorm:"TIMESTAMP"`
	Forked      int       `xorm:"comment('是否分叉成孤块') INT(20)"`
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
