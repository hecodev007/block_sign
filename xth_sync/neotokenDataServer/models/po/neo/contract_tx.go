package neo

import (
	"neotokenDataServer/common/db"
	"neotokenDataServer/common/log"
)

type ContractTx struct {
	Id       int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid     string `xorm:"not null default '' comment('交易id') index unique(idx_txid_height) VARCHAR(100)"`
	Height   int64  `xorm:"not null default 0 comment('区块高度索引值') index unique(idx_txid_height) BIGINT(20)"`
	Hash     string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	From     string `xorm:"not null default '0' VARCHAR(100)"`
	Vmstate  string `xorm:"not null default '' VARCHAR(20)"`
	To       string `xorm:"default '' VARCHAR(100)"`
	Value    int64  `xorm:"BIGINT(20)"`
	Vdecimal string `xorm:"VARCHAR(50)"`
	Contract string `xorm:"VARCHAR(45)"`
	Index    int    `xorm:"INT(11)"`
	Coinname string `xorm:"VARCHAR(45)"`
}

func ContractTxRollBack(height int64) (int64, error) {
	bl := new(ContractTx)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertContractTx(vousts []*ContractTx) (affected int64, err error) {

	affected, err = db.SyncConn.Insert(vousts)
	if err != nil {
		log.Warn(err.Error())
	}
	return
}
