package btc

import (
	"bytes"
	"github.com/shopspring/decimal"
	"strings"
	"telosDataServer/db"
	"time"
)

type BlockTX struct {
	Id          int64           `json:"id,omitempty" gorm:"column:id"`
	Txid        string          `json:"txid,omitempty" gorm:"column:txid"`
	BlockHeight int64           `json:"height,omitempty" gorm:"column:height"`
	BlockHash   string          `json:"blockhash,omitempty" gorm:"column:blockhash"`
	Version     int             `json:"version,omitempty" gorm:"column:version"`
	Size        int             `json:"size,omitempty" gorm:"column:size"`
	Fee         decimal.Decimal `json:"fee,omitempty" gorm:"column:fee"`
	VinCount    int             `json:"vincount,omitempty" gorm:"column:vincount"`
	VoutCount   int             `json:"voutcount,omitempty" gorm:"column:voutcount"`
	Coinbase    int             `json:"coinbase,omitempty" gorm:"column:coinbase"`
	Timestamp   time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime  time.Time       `json:"createtime,omitempty" gorm:"column:createtime"`
	Forked      int             `json:"forked,omitempty" gorm:"column:forked"` //是否因为分叉无效
}

func (b *BlockTX) TableName() string {
	return "block_tx"
}

func BatchInsertBlockTXs(bs []*BlockTX) error {
	num := len(bs)
	if num == 0 {
		return nil
	}

	if num <= MaxRows {
		return subBatchInsertBlockTXs(bs)
	} else {
		for start := 0; start < num; start += MaxRows {
			end := start + MaxRows
			if end > num-1 {
				end = num - 1
			}
			tmp := bs[start:end]
			if err := subBatchInsertBlockTXs(tmp); err != nil {
				return err
			}
		}
	}

	return nil
}

func subBatchInsertBlockTXs(bs []*BlockTX) error {
	num := len(bs)
	if num == 0 {
		return nil
	}

	var vals []interface{}
	for _, b := range bs {
		vals = append(vals, b.Txid, b.BlockHeight, b.BlockHash, b.Version, b.Size, b.Fee, b.VinCount, b.VoutCount, b.Coinbase, b.Timestamp, b.CreateTime, b.Forked)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx(txid,height,blockhash,version,size,fee,vincount,voutcount,coinbase,timestamp,createtime,forked) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?),`, num))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}

	return nil
}

// 删除区块
func DeleteBlockTX(height int64) error {
	err := db.SyncDB.DB.Exec("delete from block_tx where height >= ?", height).Error
	if err == nil {
		return nil
	}

	return err
}

// index 根据区块高度索引获取交易数据
func SelectBlockTXByIndex(index int64) error {
	return nil
}

// hash 获取交易数据
func SelectBlockTXByHash(hash string) (*BlockTX, error) {
	a := &BlockTX{}
	err := db.SyncDB.DB.Where("Hash = ?", hash).First(a).Error
	if err != nil {
		return nil, err
	}
	return a, nil
}

// txid 获取交易数据
func SelectBlockTX(txid string) (*BlockTX, error) {
	a := &BlockTX{}
	err := db.SyncDB.DB.Where("txid = ?", txid).First(a).Error
	if err != nil {
		return nil, err
	}
	return a, nil
}

// 插入交易数据
func InsertBlockTX(b *BlockTX) (int64, error) {
	//b.Id = int64(IDGen.Generate())

	err := db.SyncDB.DB.Create(b).Error
	if err != nil {
		return -1, nil
	}

	return b.Id, err
}
