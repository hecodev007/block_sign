package xrp

import (
	"starDataServer/common/db"
	"starDataServer/common/log"
)

type TokenTx struct {
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
	Index    int    `xorm:"default 0 unique(idx_txid_height) INT(11)"`
	Coinname string `xorm:"default '' VARCHAR(45)"`
	Memo     int64  `xorm:"BIGINT(20)"`
}

func (m *TokenTx) TableName() string {
	return "token_tx"
}

func TokenTxRollBack(height int64) (int64, error) {
	bl := new(TokenTx)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTokenTx(vousts []*TokenTx) (affected int64, err error) {

	affected, err = db.SyncConn.Insert(vousts)
	if err != nil {
		log.Warn(err.Error())
	}
	return
}
