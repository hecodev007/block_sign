package doge

import (
	"nyzoDataServer/common/db"
	"nyzoDataServer/common/log"
	"time"
)

//ALTER TABLE `dogesync`.`block_tx`
//CHANGE COLUMN `fee` `fee` VARCHAR(64) NOT NULL DEFAULT '\"0\"' ;
type BlockTx struct {
	Id         int64     `xorm:"pk autoincr BIGINT(20)"`
	Txid       string    `xorm:"not null default '' comment('txid') unique VARCHAR(100)"`
	Height     int64     `xorm:"not null comment('序号') BIGINT(20)"`
	Blockhash  string    `xorm:"default '' VARCHAR(64)"`
	Version    int       `xorm:"default 0 INT(11)"`
	Fee        string    `xorm:"default '' VARCHAR(64)"`
	Vincount   int       `xorm:"not null INT(20)"`
	Voutcount  int       `xorm:"not null INT(20)"`
	Timestamp  time.Time `xorm:"comment('æ—¶é—´æˆ³') TIMESTAMP"`
	Createtime time.Time `xorm:"comment('åˆ›å»ºæ—¶é—´') TIMESTAMP"`
	Forked     int       `xorm:"INT(20)"`
	Iscoinbase bool      `xorm:"-"`
}

func (m *BlockTx) TableName() string {
	return "block_tx"
}

func TxRollBack(height int64) (int64, error) {
	bl := new(BlockTx)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTx(tx *BlockTx) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Warn(err.Error())
	}
	return tx.Id, err
}
