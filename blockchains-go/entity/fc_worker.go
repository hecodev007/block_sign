package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/builder"
)

type FcWorker struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(10)"`
	WorkerCode string `json:"worker_code" xorm:"VARCHAR(15)"`
	CoinName   string `json:"coin_name" xorm:"VARCHAR(255)"`
	Weight     int    `json:"weight" xorm:"not null default 100 INT(11)"`
	Status     int    `json:"status" xorm:"not null default 0 comment('0禁用
1启用') TINYINT(4)"`
}

func (o *FcWorker) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcWorker) Find(cond builder.Cond) ([]*FcWorker, error) {
	res := make([]*FcWorker, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
