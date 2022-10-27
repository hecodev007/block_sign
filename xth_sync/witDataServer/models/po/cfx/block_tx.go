package cfx

import (
	"time"
	"witDataServer/common/db"
	"witDataServer/common/log"
)

type BlockTx struct {
	Id            int64     `xorm:"pk autoincr BIGINT(20)"`
	CoinName      string    `xorm:"not null comment('币种名称') VARCHAR(20)"`
	Txid          string    `xorm:"not null default '' comment('交易hash') unique VARCHAR(100)"`
	Contract      string    `xorm:"default '' VARCHAR(20)"`
	FromAddress   string    `xorm:"not null default '' comment('转出地址') index VARCHAR(100)"`
	ToAddress     string    `xorm:"not null default '' comment('接收地址') index VARCHAR(100)"`
	BlockHeight   int64     `xorm:"not null comment('高度') BIGINT(20)"`
	BlockHash     string    `xorm:"not null comment('块hash') VARCHAR(100)"`
	Amount        string    `xorm:"not null default '0' comment('金额') VARCHAR(50)"`
	Memo          string    `xorm:"not null comment('memo') VARCHAR(100)"`
	Status        int       `xorm:"comment('0代表 失败,1代表成功,2代表上链成功但交易失败') TINYINT(3)"`
	Confirmations int64     `xorm:"not null BIGINT(20)"`
	Timestamp     time.Time `xorm:"comment('交易时间戳') TIMESTAMP"`
	CreateTime    time.Time `xorm:"comment('创建时间') TIMESTAMP"`
	Fee           string    `xorm:"VARCHAR(45)"`
}

func TxRollBack(height int64) (int64, error) {
	bl := new(BlockTx)
	return db.SyncConn.Where("block_height >= ?", height).Delete(bl)
}
func InsertTx(tx *BlockTx) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Warn(err.Error())
	}
	return tx.Id, err
}
