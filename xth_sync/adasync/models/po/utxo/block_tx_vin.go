package utxo

import (
	"adasync/common/db"
	"adasync/common/log"
	"time"
)

type BlockTxVin struct {
	Id           int64     `xorm:"not null pk autoincr INT(20)"`
	Txid         string    `xorm:"not null VARCHAR(100)"`
	VoutN        int       `xorm:"not null INT(20)"`
	Blockhash    string    `xorm:"default '' VARCHAR(64)"`
	Value        string    `xorm:"not null DECIMAL(50)"`
	Address      string    `xorm:"not null VARCHAR(100)"`
	AssertId     string    `xorm:"not null VARCHAR(50)"` //ada链上代币名称,也是我们填写的代币token地址
	AssertName   string    `xorm:"not null VARCHAR(50)"` //ada我们填写的代币名称
	AssertValue  string    `xorm:"not null DECIMAL(50)"` //ada我们填写的代币额度
	Timestamp    time.Time `xorm:"TIMESTAMP"`
	SpendTxid    string    `xorm:"not null VARCHAR(100)"`
	Height       int       `xorm:"INT(20)"`
	Createtime   time.Time `xorm:"TIMESTAMP"`
	Forked       int       `xorm:"comment('是否分叉成孤块') INT(20)"`
	Scriptpubkey string    `xorm:"default '' VARCHAR(128)"`
}

func TxVinRollBack(height int64) (int64, error) {
	bl := new(BlockTxVin)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func InsertTxVin(txs []*BlockTxVin) (affected int64, err error) {
	affected, err = db.SyncConn.Insert(txs)
	if err != nil {
		log.Warn(err.Error())
	}
	return affected, err
}
