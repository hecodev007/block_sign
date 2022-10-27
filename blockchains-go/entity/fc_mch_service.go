package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/builder"
)

type FcMchService struct {
	Id        int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	MchId     int    `json:"mch_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	CoinId    int    `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	CoinName  string `json:"coin_name" xorm:"comment('币种名称') VARCHAR(20)"`
	StartTime int64  `json:"start_time" xorm:"not null default 0 comment('开通时间') INT(11)"`
	EndTime   int64  `json:"end_time" xorm:"not null default 0 comment('到期时间') INT(11)"`
	Status    int    `json:"status" xorm:"not null default 0 comment('服务状态 0 正常  1 暂停') TINYINT(3)"`
}

func (o FcMchService) Find(cond builder.Cond) ([]*FcMchService, error) {
	res := make([]*FcMchService, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
