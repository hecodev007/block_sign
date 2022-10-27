package kava

import (
	"bytes"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTX struct {
	Id          int64     `json:"id,omitempty" gorm:"column:id"`
	Txid        string    `json:"txid,omitempty" gorm:"column:txid"`
	BlockHeight int64     `json:"block_height,omitempty" gorm:"column:block_height"`
	BlockHash   string    `json:"block_hash,omitempty" gorm:"column:block_hash"`
	Fee         string    `json:"fee,omitempty" gorm:"column:fee"`
	GasUsed     int64     `json:"gasused,omitempty" gorm:"column:gasused"`
	GasWanted   int64     `json:"gaswanted,omitempty" gorm:"column:gaswanted"`
	RawLogs     string    `json:"rawlogs,omitempty" gorm:"column:rawlogs"`
	Type        string    `json:"type,omitempty" gorm:"column:type"`
	Memo        string    `json:"memo,omitempty" gorm:"column:memo"`
	MsgCount    int       `json:"msgcount,omitempty" gorm:"column:msgcount"`
	Timestamp   time.Time `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime  time.Time `json:"createtime,omitempty" gorm:"column:createtime"`
}

func (o *BlockTX) TableName() string {
	return "block_tx"
}

func BatchInsertBlockTX(bs []*BlockTX) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}

	var vals []interface{}
	for _, b := range bs {
		vals = append(vals, b.Txid, b.BlockHeight, b.BlockHash,
			b.Fee, b.GasUsed, b.GasWanted, b.RawLogs, b.Type, b.Memo, b.MsgCount, b.Timestamp, b.CreateTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx(txid,block_height,block_hash,fee,
           gasused,gaswanted,rawlogs,type,memo,msgcount,timestamp,create_time) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// 删除区块
func DeleteBlockTX(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_tx where block_height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// index 根据区块高度索引获取交易数据
func SelectBlockTXsByIndex(blkheight int64) (bs []*BlockTX, err error) {

	if err = db.SyncDB.DB.Where(" block_height = ? ", blkheight).Find(bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// hash 获取交易数据
func SelecBlockTxByHash(hash string) (*BlockTX, error) {
	b := &BlockTX{}
	if err := db.SyncDB.DB.Where(" txid = ? ", hash).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 插入交易数据
func InsertBlockTX(b *BlockTX) (int64, error) {

	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
