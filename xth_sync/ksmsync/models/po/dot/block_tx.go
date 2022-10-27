package dot

import (
	"ksmsync/common/db"
)

type BlockTx struct {
	Id              int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid            string `xorm:"not null default '' comment('交易id') index unique(idx_txid_height) VARCHAR(100)"`
	Height          int64  `xorm:"not null default 0 comment('区块高度索引值') index unique(idx_txid_height) BIGINT(20)"`
	Hash            string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	SysFee          string `xorm:"not null default 0.000000000000000000 comment('手续费') DECIMAL(40,18)"`
	Fromaccount     string `xorm:"not null default '' comment('from') VARCHAR(100)"`
	Toaccount       string `xorm:"not null default '' comment('to') VARCHAR(100)"`
	Amount          string `xorm:"not null default 0.000000000000000000 comment('金额') DECIMAL(40,18)"`
	Memo            string `xorm:"not null default '' comment('备注') VARCHAR(255)"`
	Contractaddress string `xorm:"not null default '' comment('合约地址') VARCHAR(255)"`
	//Succuss         int    `xorm:"default 1 comment('交易状态') TINYINT(1)"`
}

func (o *BlockTx) TableName() string {
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
		//log.Warn(err.Error())
	}
	return tx.Id, err
}
