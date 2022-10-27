package fil

import (
	"crustDataServer/common/db"
	"crustDataServer/common/log"
	"encoding/json"
	"time"
)

type Cids []string

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
}

func MaxBlockHeight() (h int64, err error) {
	bl := new(BlockChain)
	if _, err = db.SyncConn.Desc("height").Get(bl); err != nil {
		log.Error(err.Error())
	}
	return bl.Height, err
}
func InsertBlockChain(d *BlockChain) (int64, error) {
	_, err := db.SyncConn.InsertOne(d)
	return d.Id, err
}
