package vet

import (
	"bytes"
	"github.com/shopspring/decimal"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTxDetail struct {
	Id              int64           `json:"id,omitempty" gorm:"column:id"`
	CoinName        string          `json:"coin_name" gorm:"column:coin_name"`
	Txid            string          `json:"txid,omitempty" gorm:"column:txid"`
	Index           int             `json:"index,omitempty" gorm:"column:index"`
	BlockHeight     int64           `json:"block_height" gorm:"column:block_height"`
	BlockHash       string          `json:"block_hash,omitempty" gorm:"column:block_hash"`
	ContractAddress string          `json:"contract_address" gorm:"column:contract_address"`
	FromAddress     string          `json:"from_address" gorm:"column:from_address"`
	ToAddress       string          `json:"to_address" gorm:"column:to_address"`
	Amount          decimal.Decimal `json:"amount,omitempty" gorm:"column:amount"` //该笔交易发起人
	Status          int             `json:"status,omitempty" gorm:"column:status"`
	Timestamp       time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime      time.Time       `json:"createtime,omitempty" gorm:"column:createtime"`
}

func (o *BlockTxDetail) TableName() string {
	return "block_tx_detail"
}

func BatchInsertBlockTxDetail(bs []*BlockTxDetail) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}
	vals := make([]interface{}, 0, len(bs)*20)
	for _, b := range bs {
		vals = append(vals, b.CoinName, b.Txid, b.Index, b.BlockHeight, b.BlockHash, b.ContractAddress, b.FromAddress, b.ToAddress, b.Amount,
			b.Status, b.Timestamp, b.CreateTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx_detail(coin_name,txid,index,block_height,block_hash,contract_address,from_address,to_address, amount,status,timestamp,createtime) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// 删除区块
func DeleteBlockTxDetail(height int64) error {
	err := db.SyncDB.DB.Exec("delete from block_tx where block_height >= ?", height).Error
	if err != nil {
		return err
	}
	return nil
}

// index 根据区块高度索引获取交易数据
func SelectBlockTxDetailsByIndex(blkheight int64) (bs []*BlockTxDetail, err error) {
	if err = db.SyncDB.DB.Where(" block_height = ? ", blkheight).Find(bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// hash 获取交易数据
func SelecBlockTxDetailByHash(hash string) (bs []*BlockTxDetail, err error) {
	if err = db.SyncDB.DB.Where(" txid = ? ", hash).Find(bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// 插入交易数据
func InsertBlockTxDetail(b *BlockTxDetail) (int64, error) {
	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
