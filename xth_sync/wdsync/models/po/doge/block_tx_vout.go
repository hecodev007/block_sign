package doge

import (
	"wdsync/common/db"
	"wdsync/common/log"
	"time"
)

type BlockTxVout struct {
	Id           int64     `xorm:"not null pk autoincr INT(20)"`
	Txid         string    `xorm:"not null VARCHAR(100)"`
	Index        int       `xorm:"not null INT(20)"`
	Blockhash    string    `xorm:"default '' VARCHAR(64)"`
	Amount       string    `xorm:"not null DECIMAL(50)"`
	Address      string    `xorm:"not null VARCHAR(100)"`
	Timestamp    time.Time `xorm:"TIMESTAMP"`
	SpendTxid    string    `xorm:"not null VARCHAR(100)"`
	Height       int       `xorm:"INT(20)"`
	Createtime   time.Time `xorm:"TIMESTAMP"`
	Forked       int       `xorm:"comment('是否分叉成孤块') INT(20)"`
	Scriptpubkey string    `xorm:"default '' VARCHAR(128)"`
}

func TxVoutRollBack(height int64) (int64, error) {
	bl := new(BlockTxVout)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTxVout(tx *BlockTxVout) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Info(err.Error())
	}
	return tx.Id, err
}
