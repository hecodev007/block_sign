package btc

import (
	"bytes"
	"fmt"
	"github.com/shopspring/decimal"
	"neotokenDataServer/common/db"
	"strings"
	"time"
)

type BlockTXVout struct {
	Id           int64           `json:"id,omitempty" gorm:"column:id"`
	Txid         string          `json:"txid,omitempty" gorm:"column:txid"`
	Voutn        int             `json:"vout_n,omitempty" gorm:"column:vout_n"`
	BlockHash    string          `json:"blockhash,omitempty" gorm:"column:blockhash"`
	Value        decimal.Decimal `json:"value,omitempty" gorm:"column:value"`
	Address      string          `json:"address,omitempty" gorm:"column:address"`
	Status       int             `json:"status,omitempty" gorm:"column:status"` //0 为未花费， 1为已花费
	SpendTxid    string          `json:"spend_txid,omitempty" gorm:"column:spend_txid"`
	Timestamp    time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	ScriptPubKey string          `json:"script_pubkey,omitempty" gorm:"column:script_pubkey"`
	Coinbase     int             `json:"coinbase,omitempty" gorm:"column:coinbase"`
	CreateTime   time.Time       `json:"createtime,omitempty" gorm:"column:createtime"`
	Forked       int             `json:"forked,omitempty" gorm:"column:forked"` //是否因为分叉无效
}

func (b *BlockTXVout) TableName() string {
	return "block_tx_vout"
}

// 删除区块
func DeleteBlockTXVout(height int64) error {
	err := db.SyncDB.DB.Exec("delete from block_tx_vout where height >= ?", height).Error
	if err != nil {
		return err
	}
	return nil
}

// index 根据区块高度索引获取交易数据
func SelectBlockTXVoutByIndex(index int64) (*BlockTXVout, error) {
	return nil, nil
}

// hash 获取交易数据
func SelectBlockTXVoutsByTxid(txid string) ([]*BlockTXVout, error) {
	var bs []*BlockTXVout
	err := db.SyncDB.DB.Where("txid = ?", txid).Find(&bs).Error
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func SelectBlockTXVinsByTxid(txid string) ([]*BlockTXVout, error) {
	var bs []*BlockTXVout
	err := db.SyncDB.DB.Where("spend_txid = ?", txid).Find(&bs).Error
	if err != nil {
		return nil, err
	}
	return bs, nil
}

// txid 获取是否存在
func BlockTXVoutCount(txid string, voutn int) (int64, error) {
	var b int64
	if err := db.SyncDB.DB.Table("block_tx_vout").Where("txid = ? and vout_n = ?", txid, voutn).Count(&b).Error; err != nil {
		return -1, err
	}
	return b, nil
}

// txid 获取交易数据
func SelectBlockTXVout(txid string, voutn int) (*BlockTXVout, error) {
	b := &BlockTXVout{}
	if err := db.SyncDB.DB.Where(" txid = ? and vout_n = ? ", txid, voutn).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 插入交易数据
func InsertBlockTXVout(b *BlockTXVout) (int64, error) {
	//b.Id = int64(IDGen.Generate())
	err := db.SyncDB.DB.Create(b).Error
	if err != nil {
		return -1, err
	}
	return b.Id, nil
}

func BatchInsertBlockTXVouts(bs []*BlockTXVout) error {
	num := len(bs)
	if num == 0 {
		return nil
	}
	if num <= MaxRows {
		return subBatchInsertBlockTXVouts(bs)
	} else {
		for start := 0; start < num; start += MaxRows {
			end := start + MaxRows
			if end > num-1 {
				end = num - 1
			}
			tmp := bs[start:end]
			if err := subBatchInsertBlockTXVouts(tmp); err != nil {
				return err
			}
		}
	}
	return nil
}

func subBatchInsertBlockTXVouts(bs []*BlockTXVout) error {
	num := len(bs)
	if num == 0 {
		return nil
	}
	var vals []interface{}
	for _, b := range bs {
		vals = append(vals, b.Txid, b.Voutn, b.BlockHash, b.Value, b.Address, b.Status, b.Forked, b.SpendTxid, b.Timestamp, b.ScriptPubKey, b.CreateTime, b.Coinbase)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx_vout(txid,vout_n,blockhash,value,address,status,forked,spend_txid,timestamp,script_pubkey,createtime,coinbase) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?),`, num))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

func BatchUpdateBlockTXVouts(bs []*BlockTXVout) error {
	num := len(bs)
	if num == 0 {
		return nil
	}
	if num <= MaxRows {
		return subBatchUpdateBlockTXVouts(bs)
	} else {
		for start := 0; start < num; start += MaxRows {
			end := start + MaxRows
			if end > num-1 {
				end = num - 1
			}
			tmp := bs[start:end]
			if err := subBatchUpdateBlockTXVouts(tmp); err != nil {
				return err
			}
		}
	}
	return nil
}

func subBatchUpdateBlockTXVouts(bs []*BlockTXVout) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}

	ids := make([]int64, 0, rn)
	updateMap := make(map[string]map[int64]interface{})
	valMap1 := make(map[int64]interface{})
	valMap2 := make(map[int64]interface{})
	var vals []interface{}

	for _, b := range bs {
		ids = append(ids, b.Id)
		valMap1[b.Id] = b.SpendTxid
		valMap2[b.Id] = b.Status
	}
	updateMap["spend_txid"] = valMap1
	updateMap["status"] = valMap2

	sqlStr := bytes.NewBufferString("UPDATE block_tx_vout SET ")
	for attrName, attrValue := range updateMap {
		sqlStr.WriteString(fmt.Sprintf("%s = CASE id ", attrName))
		for bid, v := range attrValue {
			sqlStr.WriteString(fmt.Sprintf(" WHEN %d THEN ?", bid))
			vals = append(vals, v)
		}
		sqlStr.WriteString(" END,")
	}
	sqlStr.Truncate(sqlStr.Len() - 1)
	sqlStr.WriteString(" WHERE id IN (?)")
	vals = append(vals, ids)
	//log.Infof("sql: %s , rn : %d , len : %d", sqlStr.String(), rn, len(vals))

	err := db.SyncDB.DB.Exec(sqlStr.String(), vals...).Error
	if err != nil {
		return err
	}

	return nil
}

// 插入交易数据, 返回id
func UpdateBlockTXVout(b *BlockTXVout) error {
	if err := db.SyncDB.DB.Save(b).Error; err != nil {
		return err
	}
	return nil
}
