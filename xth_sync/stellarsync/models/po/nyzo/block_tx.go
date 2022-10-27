package nyzo

import (
	"stellarsync/common/db"
	"stellarsync/common/log"
	"time"
)

type BlockTx struct {
	Id          int64     `xorm:"pk autoincr BIGINT(20)"`
	BlockHeight int64     `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	Txid        string    `xorm:"not null default '' comment('交易id') unique VARCHAR(100)"`
	From        string    `xorm:"VARCHAR(100)"`
	To          string    `xorm:"VARCHAR(100)"`
	Fee         string    `xorm:"VARCHAR(100)"`
	Memo        string    `xorm:"VARCHAR(255)"`
	Type        string    `xorm:"index VARCHAR(40)"`
	BlockHash   string    `xorm:"not null default '' comment('区块hash值') VARCHAR(100)"`
	Value       string    `xorm:"VARCHAR(100)"`
	Timestamp   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('交易时间戳') TIMESTAMP"`
	Createtime  time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('创建时间') TIMESTAMP"`
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
