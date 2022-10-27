package fio

import (
	"bytes"
	"fmt"
	"rsksync/db"
	"strings"
	"time"
)

const MaxRows = 1000

type BlockInfo struct {
	Id             int64     `json:"id,omitempty" gorm:"column:id"`
	Height         int64     `json:"height,omitempty" gorm:"column:height"`
	Hash           string    `json:"hash,omitempty" gorm:"column:hash"`
	FrontBlockHash string    `json:"previousblockhash,omitempty" gorm:"column:previousblockhash"`
	NextBlockHash  string    `json:"nextblockhash,omitempty" gorm:"column:nextblockhash"`
	Timestamp      time.Time `json:"timestamp,omitempty" gorm:"column:timestamp"`
	Transactions   int       `json:"transactions,omitempty" gorm:"column:transactions"`
	Confirmations  int64     `json:"confirmations,omitempty" gorm:"column:confirmations"`
	CreateTime     time.Time `json:"createtime,omitempty" gorm:"column:createtime"`
}

func (o *BlockInfo) TableName() string {
	return "block_info"
}

func BatchInsertBlockInfo(bs []*BlockInfo) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}

	vals := make([]interface{}, 0, len(bs)*20)
	for _, b := range bs {
		vals = append(vals, b.Height, b.Hash, b.FrontBlockHash, b.NextBlockHash,
			b.Timestamp, b.Transactions, b.Confirmations, b.CreateTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_info(height,hash,previousblockhash,nextblockhash,timestamp,transactions,confirmations,createtime) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// 获取db存储最大区块高度
func GetMaxBlockIndex() (b int64, err error) {

	row := db.SyncDB.DB.Table("block_info").Select("max(height)").Row()
	if row == nil {
		return 0, fmt.Errorf("don't have max height")
	}

	if err = row.Scan(&b); err != nil {
		return 0, err
	}

	return b, nil
}

// 获取所有未确认的已推送的区块。
func GetUnconfirmBlockInfos(confirmations int64) (bs []*BlockInfo, err error) {

	if err := db.SyncDB.DB.Where(" confirmations < ? ", confirmations).Find(&bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// 批量更新确认数
func BatchUpdateConfirmations(ids []int64, sept int) error {

	err := db.SyncDB.DB.Exec("update block_info set confirmations = confirmations + (?) where id IN (?)", sept, ids).Error
	if err != nil {
		return err
	}
	return nil
}

// 更新确认数
func UpdateConfirmations(height int64, confirmations int64, nextblockhash string) error {

	err := db.SyncDB.DB.Exec("update block_info set confirmations = ?, nextblockhash = ? where height = ?", confirmations, nextblockhash, height).Error
	if err != nil {
		return err
	}
	return nil
}

// 删除区块
func DeleteBlockInfo(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_info where height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// 查找指定高度索引数量
func GetBlockCountByIndex(index int64) (int64, error) {

	var b int64
	if err := db.SyncDB.DB.Model(BlockInfo{}).Where("height = ?", index).Count(&b).Error; err != nil {
		return 0, err
	}

	return b, nil
}

// 根据高度获取块数据
func GetBlockInfoByIndex(index int64) (*BlockInfo, error) {

	b := &BlockInfo{}
	if err := db.SyncDB.DB.Where(" height = ? ", index).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 查找指定hash数量
func GetBlockCountByHash(hash string) (int64, error) {

	var b int64
	if err := db.SyncDB.DB.Model(BlockInfo{}).Where("hash = ?", hash).Count(&b).Error; err != nil {
		return -1, err
	}

	return b, nil
}

// 根据 index 和 hash 获取块数据
func GetBlockInfoByHash(hash string) (*BlockInfo, error) {

	b := &BlockInfo{}
	if err := db.SyncDB.DB.Where(" hash = ? ", hash).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 插入块数据
// return 影响行
func InsertBlockInfo(b *BlockInfo) (int64, error) {

	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
