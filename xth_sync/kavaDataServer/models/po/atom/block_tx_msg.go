package atom

import (
	"kavaDataServer/common/db"
	"kavaDataServer/common/log"
	"time"
)

type BlockTxMsg struct {
	Id          int64     `xorm:"pk autoincr BIGINT(20)"`
	CoinName    string    `xorm:"not null comment('币种名称') VARCHAR(20)"`
	Txid        string    `xorm:"not null default '' comment('交易id') unique(txidindex) VARCHAR(100)"`
	Index       int       `xorm:"not null default 0 comment('下标') unique(txidindex) INT(11)"`
	BlockHeight int64     `xorm:"not null comment('高度') BIGINT(20)"`
	BlockHash   string    `xorm:"not null comment('块hash') VARCHAR(100)"`
	FromAddress string    `xorm:"not null default '' comment('转出地址') index VARCHAR(100)"`
	ToAddress   string    `xorm:"not null default '' comment('接收地址') index VARCHAR(100)"`
	Amount      []byte    `xorm:"not null comment('金额') VARBINARY(100)"`
	Status      int       `xorm:"not null default 1 comment('0代表未成功，1代表已成功') INT(11)"`
	Timestamp   time.Time `xorm:"TIMESTAMP"`
	Createtime  time.Time `xorm:"TIMESTAMP"`
	Log         string    `xorm:"TEXT"`
	Type        string    `xorm:"VARCHAR(50)"`
}

func InsertTxMsg(tx *BlockTxMsg) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(tx)
	if err != nil {
		log.Warn(err.Error())
	}
	return tx.Id, err
}
