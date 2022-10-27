package avax

import (
	"avaxDataServer/common/log"
	"avaxDataServer/db"
	"avaxDataServer/utils/avax"
	"encoding/json"
	"errors"
	"time"
)

type BlockTx struct {
	Id         int64     `xorm:"pk autoincr BIGINT(20)"`
	Txid       string    `xorm:"not null default '' comment('txid') unique VARCHAR(100)"`
	Height     int64     `xorm:"not null comment('序号') BIGINT(20)"`
	Fee        int       `xorm:"not null default 0 INT(20)"`
	Vincount   int       `xorm:"not null INT(20)"`
	Voutcount  int       `xorm:"not null INT(20)"`
	Timestamp  time.Time `xorm:"comment('æ—¶é—´æˆ³') TIMESTAMP"`
	Createtime time.Time `xorm:"comment('åˆ›å»ºæ—¶é—´') TIMESTAMP"`
	Forked     int       `xorm:"INT(20)"`
	Rawtx string  `xorm:"text"`
}

func GetBlockTxByTxid(txid string) (*BlockTx, error) {
	tx := new(BlockTx)
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	has, err := db.SyncConn.Where("txid=?", txid).Get(tx)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("tx not find")
	}
	return tx, nil
}
func DeleteBlockInfo(h int64) error {
	return nil
}
func DeleteBlockTX(h int64) error {
	tx := new(BlockTx)
	_, err := db.SyncConn.Where("heidht >=?", h).Delete(tx)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func GetMaxBlockIndex() (int64, error) {
	tx := new(BlockTx)
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)

	_, err := db.SyncConn.OrderBy("id desc").Get(tx)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}
	return tx.Height, nil
}

func GetBlockCountByHash(hash string) (int64, error) {
	return db.SyncConn.Where("Txid=?", hash).Count(new(BlockTx))
}
func BatchInsertBlockTXs(txs []*avax.Transaction, height int64) error {
	blocktxs := make([]*BlockTx, 0, len(txs))
	for _, v := range txs {
		tx := new(BlockTx)
		tx.Height = height
		tx.Txid = v.ID
		tx.Createtime = v.Timestamp
		tx.Vincount = len(v.Inputs)
		tx.Voutcount = len(v.Outputs)
		tx.Timestamp = time.Now()

		rawtx,_ := json.Marshal(v)
		tx.Rawtx = string(rawtx)
		blocktxs = append(blocktxs, tx)
	}
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	_, err := db.SyncConn.Insert(blocktxs)
	return err
}
