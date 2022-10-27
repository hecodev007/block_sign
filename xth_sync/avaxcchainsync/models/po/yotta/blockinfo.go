package yotta

import (
	"avaxcchainsync/common/db"
	"avaxcchainsync/common/log"
	"time"
)

const MaxRows = 1000

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
}

func (o *BlockInfo) TableName() string {
	return "block_info"
}

func MaxBlockHeight() (h int64, err error) {
	bl := new(BlockInfo)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Warn(err.Error())
	}
	return bl.Height, err
}

func BlockHashExist(blockHash string) (bool, error) {
	bl := new(BlockInfo)
	return db.SyncConn.Where("hash=?", blockHash).Exist(bl)
}

// 删除区块
func BlockRollBack(height int64) (int64, error) {
	bl := new(BlockInfo)
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	ret, err := db.SyncConn.Exec("delete from "+bl.TableName()+" where height >= ?", height)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// 插入块数据
// return 影响行
func InsertBlock(block *BlockInfo) (int64, error) {
	_, err := db.SyncConn.InsertOne(block)
	return block.Id, err
}
