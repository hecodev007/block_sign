package crust

import (
	"crustDataServer/common/db"
	"crustDataServer/common/log"
)

type BlockTx struct {
	Id          int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid        string `xorm:"not null default '' comment('交易id') index VARCHAR(100)"`
	Height      int64  `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	Hash        string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	Fee         string `xorm:"not null default '0' comment('手续费') VARCHAR(64)"`
	Fromaccount string `xorm:"not null default '' comment('from') VARCHAR(100)"`
	Toaccount   string `xorm:"not null default '' comment('to') VARCHAR(100)"`
	Amount      string `xorm:"not null default '0' comment('金额') VARCHAR(64)"`
	Memo        string `xorm:"not null default '' comment('备注') VARCHAR(255)"`
	Status      int    `xorm:"not null default 1 comment('交易状态,1成功') INT(11)"`
}

func (m *BlockTx) TableName() string {
	return "block_tx"
}
func TxRollBack(height int64) (int64, error) {
	bl := new(BlockTx)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTx(tx *BlockTx) (id int64, err error) {
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Info(err.Error())
	}
	return tx.Id, err
}
