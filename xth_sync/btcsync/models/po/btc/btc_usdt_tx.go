package btc

import (
	"btcsync/common/db"
	"btcsync/common/log"

	"github.com/shopspring/decimal"
)

type BtcUsdtTx struct {
	Id               int64           `gorm:"column:id"`
	Txid             string          `gorm:"column:txid"`
	Fee              decimal.Decimal `gorm:"column:fee"`
	Sendingaddress   string          `gorm:"column:sendingaddress"`
	Referenceaddress string          `gorm:"column:referenceaddress"`
	Amount           decimal.Decimal `gorm:"column:amount"`
	Valid            bool            `gorm:"column:valid"`
	Blockhash        string          `gorm:"column:blockhash"`
	Blocktime        int64           `gorm:"column:blocktime"`
	Block            int64           `gorm:"column:block"`
}

func (b *BtcUsdtTx) TableName() string {
	return "block_usdt_tx"
}

func InsertBtcUsdtTx(txs []*BtcUsdtTx) (affeced int64, err error) {
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	affeced, err = db.SyncConn.Insert(txs)
	if err != nil {
		log.Warn(err.Error())
	}
	return affeced, err
}
