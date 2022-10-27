package fil

import (
	"encoding/json"
	"errors"
	"filDataServer/common/db"
	"filDataServer/common/log"
	"time"
)

type Cids []map[string]string

func (c *Cids) FromDB(bytes []byte) error {
	if len(bytes) == 0 {
		bytes = []byte("[]")
	}
	return json.Unmarshal(bytes, c)
}
func (c *Cids) ToDB() (bytes []byte, err error) {
	bytes, err = json.Marshal(c)
	return
}

type BlockChain struct {
	Id            int64     `xorm:"pk BIGINT(20)"`
	Height        int64     `xorm:"not null default 0 BIGINT(20)"`
	Confirmations int64     `xorm:"comment('确认数') INT(11)"`
	Blocknum      int       `xorm:"not null default 0 INT(11)"`
	Timestamp     time.Time `xorm:"TIMESTAMP(6)"`
	Cids          Cids      `xorm:"TEXT"`
	Parent        Cids      `xorm:"TEXT"`
	Exed          int       `xorm:"INT(11)"`
	//Parentbasefee decimal.Decimal `xorm:"decimal(50,0)"`
	Parentbasefee string `xorm:"default 0 DECIMAL(50)"`
}
func BlockChainRollBack(height int64) (int64, error) {
	bl := new(BlockChain)
	return db.SyncConn.Where("height >= ?", height).Delete(bl)
}
func MaxBlockHeight() (h int64, err error) {
	bl := new(BlockChain)
	if _, err = db.SyncConn.Where("exed=1").Desc("height").Get(bl); err != nil {
		log.Error(err.Error())
	}
	return bl.Height, err
}
func MaxBlockHeightNotExed() (h int64, err error) {
	bl := new(BlockChain)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Error(err.Error())
	}
	return bl.Height, err
}
func MinBlockHeight() (h int64, err error) {
	bl := new(BlockChain)
	if _, err = db.SyncConn.Asc("height").Get(bl); err != nil {
		log.Error(err.Error())
	}
	return bl.Height, err
}
func GetBlockChain(h int64) (d *BlockChain, err error) {
	d = new(BlockChain)
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	has, err := db.SyncConn.Where("id=?", h).Get(d)
	if !has {
		return d, errors.New("data not found")
	}
	return d, err
}
func CountBlock(start, end int64) (num int64) {
	d := new(BlockChain)
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	num, _ = db.SyncConn.Where("id>=? and id<=?", start, end).Count(d)
	return
}
func InsertBlockChain(d *BlockChain) (int64, error) {
	_, err := db.SyncConn.InsertOne(d)
	if err != nil {
		_, err = db.SyncConn.Where("id=?", d.Height).Update(d)
	}
	return d.Id, err
}

func ExedRollBack(height int64)  {
	db.SyncConn.ShowSQL(true)
	defer db.SyncConn.ShowSQL(false)
	bl := new(BlockChain)
	db.SyncConn.Where("id>?",height-100).Delete(bl)
	log.Info("ExedRollBack")
	//log.Info(ret.RowsAffected())
}