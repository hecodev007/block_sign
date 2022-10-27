package fil

import (
	"crustDataServer/common/db"
	"crustDataServer/common/log"
	"time"
)

type BlockTx struct {
	Id          int64     `xorm:"pk autoincr BIGINT(20)"`
	Txid        string    `xorm:"not null default '' comment('交易hash') unique VARCHAR(100)"`
	FromAddress string    `xorm:"not null default '' comment('转出地址') index VARCHAR(100)"`
	ToAddress   string    `xorm:"not null default '' comment('接收地址') index VARCHAR(100)"`
	BlockHeight int64     `xorm:"not null comment('高度') BIGINT(20)"`
	BlockHash   string    `xorm:"not null comment('块hash') VARCHAR(100)"`
	Amount      int64     `xorm:"not null comment('金额') BIGINT(20)"`
	Decimalmnt  string    `xorm:"default '' VARCHAR(50)"`
	Status      string    `xorm:"default '' comment('0代表 失败,1代表成功,2代表上链成功但交易失败') VARCHAR(64)"`
	Gaslimit    int64     `xorm:"not null BIGINT(20)"`
	Gasfeecap   int64     `xorm:"not null BIGINT(20)"`
	Gaspremium  int64     `xorm:"default 0 BIGINT(20)"`
	Nonce       int64     `xorm:"not null BIGINT(20)"`
	Params      string    `xorm:"TEXT"`
	Timestamp   time.Time `xorm:"default 'CURRENT_TIMESTAMP' comment('交易时间戳') TIMESTAMP"`
	Createtime  time.Time `xorm:"default 'CURRENT_TIMESTAMP' comment('创建时间') TIMESTAMP"`
	Method      int64     `xorm:"default 0 BIGINT(20)"`
	Version     int64     `xorm:"default 0 BIGINT(20)"`
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
func InsertTxs(tx []*BlockTx) (affected int64, err error) {
	affected, err = db.SyncConn.Insert(tx)
	if err != nil {
		log.Info(err.Error())
	}
	return affected, err
}
