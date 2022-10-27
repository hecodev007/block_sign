package doge

import (
	"kavaDataServer/common/db"
	"kavaDataServer/common/log"
	"time"
)

type BlockInfo struct {
	Id     int64  `xorm:"pk autoincr BIGINT(20)"`
	Height int64  `xorm:"not null default 0 comment('é«˜åº¦') index BIGINT(20)"`
	Hash   string `xorm:"not null default '' comment('hash') unique VARCHAR(100)"`
	//Version           int       `xorm:"INT(20)"`
	Previousblockhash string    `xorm:"not null default '' comment('ä¸Šä¸€ä¸ªåŒºå—hash') VARCHAR(100)"`
	Nextblockhash     string    `xorm:"comment('ä¸‹ä¸€ä¸ªåŒºå—hash') VARCHAR(100)"`
	Timestamp         time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('æ—¶é—´æˆ³') TIMESTAMP"`
	Transactions      int       `xorm:"not null default 0 comment('äº¤æ˜“æ•°é‡') INT(11)"`
	Confirmations     int64     `xorm:"comment('ç¡®è®¤æ•°') INT(11)"`
	Createtime        time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' comment('åˆ›å»ºæ—¶é—´') TIMESTAMP"`
	Forked            int       `xorm:"INT(11)"`
}

func MaxBlockHeight() (h int64, err error) {
	bl := new(BlockInfo)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Warn(err.Error())
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
