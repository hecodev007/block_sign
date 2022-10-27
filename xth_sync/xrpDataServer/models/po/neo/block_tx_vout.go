package neo

import (
	"xrpDataServer/common/db"
	"xrpDataServer/common/log"
)

type BlockTxVout struct {
	Id            int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid          string `xorm:"not null default '' comment('交易id') index(txid) VARCHAR(100)"`
	Height        int64  `xorm:"not null default 0 comment('区块高度索引值') index BIGINT(20)"`
	Hash          string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	VoutN         int    `xorm:"not null default 0 comment('vout索引') index(txid) INT(11)"`
	VoutValue     string `xorm:"not null default '0' comment('值') VARCHAR(100)"`
	VoutAddress   string `xorm:"not null default '' comment('收款地址(签名数量为1有效)') index VARCHAR(100)"`
	Status        int    `xorm:"not null default 1 comment('状态(1:未花费，2：已转出)') INT(11)"`
	Invaild       int    `xorm:"not null default 0 comment('是否无效(0: 有效，1无效)') INT(11)"`
	AssetName     string `xorm:"not null default '' comment('代币名称') VARCHAR(100)"`
	AssetSelltxid string `xorm:"not null default '' comment('代币发现txid') VARCHAR(255)"`
	AssetId       string `xorm:"not null default '' comment('代币id') VARCHAR(100)"`
	AssetValue    int64  `xorm:"not null default 0 comment('代币金额') BIGINT(20)"`
}

func VoutRollBack(height int64) (int64, error) {
	bl := new(BlockTxVout)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}

func InsertVouts(vousts []*BlockTxVout) (affected int64, err error) {
	affected, err = db.SyncConn.Insert(vousts)
	if err != nil {
		log.Info(err.Error())
	}
	return
}
