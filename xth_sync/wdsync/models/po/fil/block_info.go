package fil

import (
	"wdsync/common/db"
	"time"
)

type BlockInfo struct {
	Id                int64     `xorm:"pk autoincr BIGINT(20)"`
	Height            int64     `xorm:"not null default 0 comment('区块高度') index BIGINT(20)"`
	Hash              string    `xorm:"not null default '' comment('区块hash值') unique VARCHAR(100)"`
	Previousblockhash string    `xorm:"not null default '' comment('前一个区块hash') VARCHAR(100)"`
	Nextblockhash     string    `xorm:"comment('后一个区块hash') VARCHAR(100)"`
	Timestamp         time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('时间戳') TIMESTAMP"`
	Transactions      int       `xorm:"not null default 0 comment('交易总数') INT(11)"`
	Confirmations     int64     `xorm:"comment('确认数') INT(11)"`
	Createtime        time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('记录时间') TIMESTAMP"`
	Index             int       `xorm:"not null default 0 INT(11)"`
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
