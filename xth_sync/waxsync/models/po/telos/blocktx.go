package telos

import (
	"waxsync/common/db"
	"time"

	"github.com/shopspring/decimal"
)

type BlockTx struct {
	Id              int64           `xorm:"pk autoincr BIGINT(20)"`
	CoinName        string          `xorm:"default '' VARCHAR(45)"`
	Txid            string          `xorm:"not null default '' comment('交易id') unique VARCHAR(100)"`
	ContractAddress string          `xorm:"VARCHAR(45)"`
	FromAddress     string          `xorm:"default '' VARCHAR(45)"`
	ToAddress       string          `xorm:"default '' VARCHAR(45)"`
	BlockHeight     int64           `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	BlockHash       string          `xorm:"not null default '' comment('区块hash值') VARCHAR(100)"`
	Amount          decimal.Decimal `xorm:"DECIMAL(20,18)"`
	Memo            string          `xorm:"VARCHAR(255)"`
	Status          string          `xorm:"index VARCHAR(40)"`
	Timestamp       time.Time       `xorm:"not null default 'CURRENT_TIMESTAMP' comment('交易时间戳') TIMESTAMP"`
	Createtime      time.Time       `xorm:"not null default 'CURRENT_TIMESTAMP' comment('创建时间') TIMESTAMP"`
	Txjson          string          `xorm:"not null TEXT"`
}

func (o *BlockTx) TableName() string {
	return "block_tx"
}

// 删除区块
func DeleteBlockTX(height int64) error {
	err := db.SyncDB.DB.Exec("delete from block_tx where block_height >= ?", height).Error
	if err != nil {
		return err
	}
	return nil
}

// 插入交易数据
func InsertBlockTX(b *BlockTx) (int64, error) {

	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
