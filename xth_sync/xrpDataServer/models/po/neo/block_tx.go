package neo

import (
	"xrpDataServer/common/db"
	"xrpDataServer/common/log"
)

type BlockTx struct {
	Id        int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid      string `xorm:"not null default '' comment('交易id') index unique(idx_txid_height) VARCHAR(100)"`
	Height    int64  `xorm:"not null default 0 comment('区块高度索引值') index unique(idx_txid_height) BIGINT(20)"`
	Hash      string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	SysFee    int64  `xorm:"not null default 0.000000 comment('手续费') BIGINT(20)"`
	Vincount  int    `xorm:"not null default 0 INT(11)"`
	Voutcount int    `xorm:"not null default 0 INT(11)"`
	Type      string `xorm:"not null default '' VARCHAR(20)"`
	Vmstate   string `xorm:"not null default '' VARCHAR(20)"`
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
