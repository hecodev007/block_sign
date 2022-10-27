package iota

import (
	"iotasync/common/db"
	"iotasync/common/log"
	"time"
)

type BlockTxVout struct {
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

func TxVoutRollBack(height int64) (int64, error) {
	bl := new(BlockTxVout)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTxVout(txs []*BlockTxVout) (affeced int64, err error) {
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	affeced, err = db.SyncConn.Insert(txs)
	if err != nil {
		log.Warn(err.Error())
	}
	return affeced, err
}
