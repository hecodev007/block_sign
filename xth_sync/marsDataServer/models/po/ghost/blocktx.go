package ghost

import (
	"marsDataServer/common/db"
	"marsDataServer/common/log"
	"time"
)

type BlockTx struct {
	Id         int64     `xorm:"pk autoincr BIGINT(20)"`
	Txid       string    `xorm:"not null default '' comment('txid') unique VARCHAR(100)"`
	Height     int64     `xorm:"not null comment('åŒºå—é«˜åº¦') BIGINT(20)"`
	Blockhash  string    `xorm:"not null comment('åŒºå—hash') VARCHAR(100)"`
	Version    int       `xorm:"not null INT(20)"`
	Size       int       `xorm:"INT(20)"`
	Fee        string    `xorm:"not null DECIMAL(50,8)"`
	Vincount   int       `xorm:"not null INT(20)"`
	Voutcount  int       `xorm:"not null INT(20)"`
	Coinbase   int       `xorm:"not null INT(20)"`
	Timestamp  time.Time `xorm:"comment('æ—¶é—´æˆ³') TIMESTAMP"`
	Createtime time.Time `xorm:"comment('åˆ›å»ºæ—¶é—´') TIMESTAMP"`
	Forked     int       `xorm:"INT(20)"`
}

func (b *BlockTx) TableName() string {
	return "block_tx"
}
func TxRollBack(height int64) (int64, error) {
	bl := new(BlockTx)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTx(tx *BlockTx) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Info(err.Error())
	}
	return tx.Id, err
}
