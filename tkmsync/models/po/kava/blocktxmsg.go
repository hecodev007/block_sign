package kava

import (
	"bytes"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTXMsg struct {
	Id          int64     `json:"id,omitempty" gorm:"column:id"`
	CoinName    string    `json:"coin_name" gorm:"column:coin_name"`
	Txid        string    `json:"txid,omitempty" gorm:"column:txid"`
	Index       int       `json:"index,omitempty" gorm:"column:index"`
	BlockHeight int64     `json:"block_height" gorm:"column:block_height"`
	BlockHash   string    `json:"block_hash,omitempty" gorm:"column:block_hash"`
	FromAddress string    `json:"from_address,omitempty" gorm:"column:from_address"`
	ToAddress   string    `json:"to_address,omitempty" gorm:"column:to_address"`
	Amount      string    `json:"amount,omitempty" gorm:"column:amount"`
	Status      int       `json:"status,omitempty" gorm:"column:status"` //0代表未成功，1代表已成功
	Log         string    `json:"log,omitempty" gorm:"column:log"`
	Type        string    `json:"type,omitempty" gorm:"column:type"`
	Timestamp   time.Time `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime  time.Time `json:"createtime,omitempty" gorm:"column:createtime"`
	UnlockTime  string    `json:"unlock_time,omitempty" gorm:"column:unlock_time"`
}

func (b *BlockTXMsg) TableName() string {
	return "block_tx_msg"
}

func BatchInsertBlockTXMsgs(bs []*BlockTXMsg) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}
	var vals []interface{}
	for _, b := range bs {
		vals = append(vals, b.CoinName, b.Txid, b.Index, b.BlockHeight, b.BlockHash, b.FromAddress, b.ToAddress, b.Amount, b.Status, b.Log, b.Type, b.Timestamp, b.CreateTime, b.UnlockTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx_msg(coin_name,txid,index,block_height,block_hash,from_address,to_address,amount,status,log,type,timestamp,createtime,unlock_time) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// 删除区块
func DeleteBlockTXVout(height int64) error {
	err := db.SyncDB.DB.Exec("delete from block_tx_msg where block_height >= ?", height).Error
	if err != nil {
		return err
	}
	return nil
}

// hash 获取交易数据
func SelectBlockTXMsgsByTxid(txid string) ([]*BlockTXMsg, error) {
	var bs []*BlockTXMsg
	err := db.SyncDB.DB.Where("txid = ?", txid).Find(&bs).Error
	if err != nil {
		return nil, err
	}
	return bs, nil
}

// txid 获取是否存在
func BlockTXMsgCount(txid string, index int) (int64, error) {
	var b int64
	if err := db.SyncDB.DB.Table("block_tx_msg").Where("txid = ? and msg_index = ?", txid, index).Count(&b).Error; err != nil {
		return -1, err
	}
	return b, nil
}

// txid 获取交易数据
func SelectBlockTXMsg(txid string, index int) (*BlockTXMsg, error) {
	b := &BlockTXMsg{}
	if err := db.SyncDB.DB.Where(" txid = ? and msg_index = ? ", txid, index).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 插入交易数据
func InsertBlockTXMsg(b *BlockTXMsg) (int64, error) {
	err := db.SyncDB.DB.Create(b).Error
	if err != nil {
		return -1, err
	}
	return b.Id, nil
}
