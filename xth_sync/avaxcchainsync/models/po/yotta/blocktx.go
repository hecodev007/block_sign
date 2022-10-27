package yotta

import (
	"avaxcchainsync/common/db"
	"avaxcchainsync/common/log"
	"time"

	"github.com/shopspring/decimal"
)

type BlockTx struct {
	Id              int64           `xorm:"pk autoincr BIGINT(20)"`
	CoinName        string          `xorm:"default '' VARCHAR(45)"`
	Txid            string          `xorm:"not null default '' comment('交易id') unique VARCHAR(100)"`
	ContractAddress string          `xorm:"VARCHAR(45)"`
	FromAddress     string          `xorm:"default '' VARCHAR(45)"`
	ToAddress       string          `xorm:"default '' VARCHAR(45)"`
	BlockHeight     int64           `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	BlockHash       string          `xorm:"not null default '' comment('区块hash值') VARCHAR(100)"`
	Amount          decimal.Decimal `xorm:"DECIMAL(20,18)"`
	Fee             decimal.Decimal `xorm:"DECIMAL(20,18)"`
	Memo            string          `xorm:"VARCHAR(255)"`
	Status          string          `xorm:"index VARCHAR(40)"`
	Timestamp       time.Time       `xorm:"not null default 'CURRENT_TIMESTAMP' comment('交易时间戳') TIMESTAMP"`
}

func (o *BlockTx) TableName() string {
	return "block_tx"
}

func TxRollBack(height int64) (int64, error) {
	bl := new(BlockTx)
	ret, err := db.SyncConn.Exec("delete from "+bl.TableName()+" where block_height >= ?", height)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}
func InsertTx(tx *BlockTx) (id int64, err error) {
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Warn(err.Error())
	}
	return tx.Id, err
}
