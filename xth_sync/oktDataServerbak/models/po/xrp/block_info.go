package xrp

import (
	"oktDataServer/common/db"
	"oktDataServer/common/log"
	"time"
)

type BlockInfo struct {
	Id                int64     `xorm:"pk autoincr unique BIGINT(20)"`
	Height            int64     `xorm:"not null default 0 comment('区块高度索引值') unique BIGINT(20)"`
	Hash              string    `xorm:"not null default '' comment('区块hash值') unique VARCHAR(100)"`
	Previousblockhash string    `xorm:"not null default '' VARCHAR(100)"`
	Confirmations     int64     `xorm:"not null default 0 BIGINT(20)"`
	Nextblockhash     string    `xorm:"not null default '' VARCHAR(100)"`
	Time              time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('时间') TIMESTAMP"`
	Transactions      int       `xorm:"not null default 0 comment('交易总数') INT(11)"`
}

func (m *BlockInfo) TableName() string {
	return "block_info"
}

func MaxBlockHeight() (h int64, err error) {
	bl := new(BlockInfo)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Error(err.Error())
	}
	return bl.Height, err
}

func BlockRollBack(height int64) (int64, error) {
	bl := new(BlockInfo)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}

func BlockHashExist(blockHash string) (bool, error) {
	bl := new(BlockInfo)
	return db.SyncConn.Where("hash=?", blockHash).Exist(bl)
}

func InsertBlock(block *BlockInfo) (int64, error) {
	_, err := db.SyncConn.InsertOne(block)
	return block.Id, err
}
