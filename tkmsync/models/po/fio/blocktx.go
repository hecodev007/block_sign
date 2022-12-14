package fio

import (
	"bytes"
	"github.com/shopspring/decimal"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTX struct {
	Id              int64           `json:"id,omitempty" gorm:"column:id"`
	CoinName        string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	Txid            string          `json:"txid,omitempty" gorm:"column:txid"`
	ContractAddress string          `json:"contract_address,omitempty" gorm:"column:contract_address"`
	FromAddress     string          `json:"from_address,omitempty" gorm:"column:from_address"`
	ToAddress       string          `json:"to_address,omitempty" gorm:"column:to_address"`
	BlockHeight     int64           `json:"block_height,omitempty" gorm:"column:block_height"`
	BlockHash       string          `json:"block_hash,omitempty" gorm:"column:block_hash"`
	Amount          decimal.Decimal `json:"amount,omitempty" gorm:"column:amount"`
	Fee             decimal.Decimal `json:"fee,omitempty" gorm:"column:fee"`
	Memo            string          `json:"memo,omitempty" gorm:"column:memo"`
	Status          string          `json:"status,omitempty" gorm:"column:status"`
	Timestamp       time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime      time.Time       `json:"createtime,omitempty" gorm:"column:createtime"`
	TxJson          string          `json:"txjson,omitempty" gorm:"column:txjson"`
}

func (o *BlockTX) TableName() string {
	return "block_tx"
}

func BatchInsertBlockTX(bs []*BlockTX) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}

	vals := make([]interface{}, 0, len(bs)*20)
	for _, b := range bs {
		vals = append(vals, b.CoinName, b.Txid, b.ContractAddress, b.FromAddress, b.ToAddress, b.BlockHeight, b.BlockHash,
			b.Amount, b.Memo, b.Status, b.Timestamp, b.CreateTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx(coin_name,txid,contract_address,from_address,to_address,block_height,block_hash,amount,memo,status,timestamp,createtime) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// ????????????
func DeleteBlockTX(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_tx where block_height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// index ??????????????????????????????????????????
func SelectBlockTXsByIndex(blkheight int64) (bs []*BlockTX, err error) {

	if err = db.SyncDB.DB.Where(" block_height = ? ", blkheight).Find(bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// hash ??????????????????
func SelecBlockTxByHash(hash string) (*BlockTX, error) {
	b := &BlockTX{}
	if err := db.SyncDB.DB.Where(" txid = ? ", hash).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// ??????????????????
func InsertBlockTX(b *BlockTX) (int64, error) {
	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}

func encodeData() {

}
