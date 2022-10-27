package vet

import (
	"bytes"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTX struct {
	Id           int64     `json:"id,omitempty" gorm:"column:id"`
	Txid         string    `json:"txid,omitempty" gorm:"column:txid"`
	BlockHeight  int64     `json:"block_height,omitempty" gorm:"column:block_height"`
	BlockHash    string    `json:"block_hash,omitempty" gorm:"column:block_hash"`
	Origin       string    `json:"origin,omitempty" gorm:"column:origin"` //该笔交易发起人
	Status       int       `json:"status,omitempty" gorm:"column:status"`
	GasUsed      uint64    `json:"gasused,omitempty" gorm:"column:gasused"`
	GasPriceCoef int       `json:"gaspricecoef,omitempty" gorm:"column:gaspricecoef"`
	PaidVTHO     string    `json:"paid_vtho,omitempty" gorm:"column:paid_vtho"`
	RewardVTHO   string    `json:"reward_vtho,omitempty" gorm:"column:reward_vtho"`
	Nonce        string    `json:"nonce,omitempty" gorm:"column:nonce"`
	TxCount      int       `json:"txcount,omitempty" gorm:"column:txcount"`
	Timestamp    time.Time `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime   time.Time `json:"createtime,omitempty" gorm:"column:createtime"`
	ChainTag     int       `json:"chaintag,omitempty" gorm:"column:chaintag"`
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
		vals = append(vals, b.Txid, b.BlockHeight, b.BlockHash, b.Origin,
			b.Status, b.GasUsed, b.GasPriceCoef, b.Nonce, b.TxCount, b.Timestamp, b.CreateTime, b.ChainTag)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx(txid,block_height,block_hash,origin,status,gasused,gaspricecoef,paid_vtho,reward_vtho,nonce,txcount,timestamp,createtime,chaintag) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?,?,?),`, rn))
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
