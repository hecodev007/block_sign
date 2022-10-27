package neo

import (
	"neotokenDataServer/common/db"
	"neotokenDataServer/common/log"
)

type BlockTxVin struct {
	Id           int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid         string `xorm:"not null default '' comment('交易id') index VARCHAR(100)"`
	Height       int64  `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	Hash         string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	VinTxid      string `xorm:"not null default '' comment('vin交易id') VARCHAR(100)"`
	VinVoutindex int    `xorm:"not null default 0 comment('vin对应vout索引') INT(11)"`
	Address      string `xorm:"VARCHAR(45)"`
	Amount       string `xorm:"DECIMAL(50,8)"`
	Assetname    string `xorm:"VARCHAR(45)"`
	Asset        string `xorm:"VARCHAR(45)"`
}

func VinRollBack(height int64) (int64, error) {
	bl := new(BlockTxVin)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}

func InsertVins(vins []*BlockTxVin) (affected int64, err error) {
	if len(vins) == 0 {
		return 0, nil
	}
	affected, err = db.SyncConn.Insert(vins)
	if err != nil {
		log.Info(err.Error())
	}
	return
}
