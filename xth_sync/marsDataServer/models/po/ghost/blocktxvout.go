package ghost

import (
	"marsDataServer/common/log"
	"marsDataServer/common/db"
	"time"
)

type BlockTxVout struct {
	Id           int       `xorm:"not null pk autoincr INT(20)"`
	Txid         string    `xorm:"not null VARCHAR(100)"`
	VoutN        int       `xorm:"not null INT(20)"`
	Blockhash    string    `xorm:"not null VARCHAR(100)"`
	Value        string    `xorm:"not null DECIMAL(50,8)"`
	Address      string    `xorm:"not null VARCHAR(100)"`
	Status       int       `xorm:"not null INT(20)"`
	SpendTxid    string    `xorm:"not null VARCHAR(100)"`
	Timestamp    time.Time `xorm:"TIMESTAMP"`
	ScriptPubkey string    `xorm:"not null VARCHAR(512)"`
	Coinbase     int       `xorm:"INT(20)"`
	Createtime   time.Time `xorm:"TIMESTAMP"`
	Forked       int       `xorm:"INT(20)"`
}

func (b *BlockTxVout) TableName() string {
	return "block_tx_vout"
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
