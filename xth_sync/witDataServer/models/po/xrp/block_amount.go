package xrp

import (
	"time"
	"witDataServer/common/db"
)

type BlockAmount struct {
	Id          int64     `xorm:"pk autoincr BIGINT(20)"`
	Address     string    `xorm:"not null default '' comment('账号地址') index VARCHAR(100)"`
	Amount      int64     `xorm:"not null default 0 comment('金额') BIGINT(20)"`
	BlockHeight int64     `xorm:"default 0 BIGINT(20)"`
	Decimalmnt  string    `xorm:"default '' VARCHAR(45)"`
	Update      time.Time `xorm:"TIMESTAMP(6)"`
}

func (m *BlockAmount) TableName() string {
	return "block_amount"
}

func UpdateAmount(addr string, Amount int64, decimalValue string, blockHeight int64) error {
	ba := &BlockAmount{
		Address:     addr,
		Amount:      Amount,
		BlockHeight: blockHeight,
		Decimalmnt:  decimalValue,
		Update:      time.Now(),
	}
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	affected, err := db.SyncConn.Where("address=?", addr).Update(ba)
	if err != nil {
		return err
	}
	if affected == 0 {
		_, err = db.SyncConn.InsertOne(ba)
	}
	return err

}
