package xrp

import (
	"cfxDataServer/common/db"
	"cfxDataServer/common/log"
)

type BlockTx struct {
	Id       int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid     string `xorm:"not null default '' comment('交易id') index unique(idx_txid_height) VARCHAR(100)"`
	Height   int64  `xorm:"not null default 0 comment('区块高度索引值') index unique(idx_txid_height) BIGINT(20)"`
	Hash     string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	Fee      int64  `xorm:"not null default 0 comment('手续费') BIGINT(20)"`
	From     string `xorm:"not null default '' comment('from') VARCHAR(100)"`
	To       string `xorm:"not null default '' comment('to') VARCHAR(100)"`
	Amount   int64  `xorm:"not null default 0 comment('金额') BIGINT(20)"`
	Memo     int64  `xorm:"not null default '' comment('备注') VARCHAR(255)"`
	Vincount int    `xorm:"INT(11)"`
	State    string `xorm:"default '' VARCHAR(45)"`
	Type     string `xorm:"default '' VARCHAR(45)"`
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
		log.Info(err.Error())
	}
	return tx.Id, err
}
