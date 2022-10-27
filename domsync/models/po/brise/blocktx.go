package brise

import (
	"domsync/common/db"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

type BlockTx struct {
	Id              int64           `json:"id,omitempty" gorm:"column:id"`
	CoinName        string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	Txid            string          `json:"txid,omitempty" gorm:"column:txid"`
	ContractAddress string          `json:"contract_address,omitempty" gorm:"column:contract_address"`
	FromAddress     string          `json:"from_address,omitempty" gorm:"column:from_address"`
	ToAddress       string          `json:"to_address,omitempty" gorm:"column:to_address"`
	BlockHeight     int64           `json:"block_height,omitempty" gorm:"column:block_height"`
	BlockHash       string          `json:"block_hash,omitempty" gorm:"column:block_hash"`
	Amount          decimal.Decimal `json:"amount,omitempty" gorm:"column:amount"`
	Status          int             `json:"status,omitempty" gorm:"column:status"`
	GasUsed         int64           `json:"gas_used,omitempty" gorm:"column:gas_used"`
	GasPrice        int64           `json:"gas_price,omitempty" gorm:"column:gas_price"`
	Nonce           int             `json:"nonce,omitempty" gorm:"column:nonce"`
	Input           string          `json:"input,omitempty" gorm:"column:input"`
	Decimal         int             `json:"decimal,omitempty" gorm:"column:decimal"`
	Logs            string          `json:"logs,omitempty" gorm:"column:logs"`
	Timestamp       time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime      time.Time       `json:"create_time,omitempty" gorm:"column:create_time"`
	//ToAmount        decimal.Decimal `json:"toAmount" gorm:"-"` //sta这个币种会销毁币种.临时加结构处理
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
		log.Printf(err.Error())
	}
	return tx.Id, err
}
