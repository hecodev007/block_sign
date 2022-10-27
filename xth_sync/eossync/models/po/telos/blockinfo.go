package telos

import (
	"eossync/common/db"
	"eossync/common/log"
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

// 获取db存储最大区块高度
func MaxBlockHeight() (b int64, err error) {
	bl := new(BlockInfo)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Warn(err.Error())
	}
	return bl.Height, err
}

// 删除区块
func DeleteBlockInfo(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_info where height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// 查找指定hash数量
func GetBlockCountByHash(hash string) (int64, error) {

	var b int64
	if err := db.SyncDB.DB.Model(BlockInfo{}).Where("hash = ?", hash).Count(&b).Error; err != nil {
		return -1, err
	}

	return b, nil
}

// 插入块数据

// 插入块数据
func InsertBlockInfo(b *BlockInfo) (int64, error) {
	b.Id = 1
	db.SyncDB.DB.First(new(BlockInfo))
	tdb := db.SyncDB.DB.Save(b)
	if tdb.RowsAffected == 0 {
		if err := db.SyncDB.DB.Create(b).Error; err != nil {
			return 0, err
		}
	}
	return b.Id, nil
}
