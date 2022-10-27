package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
	"xorm.io/builder"
)

type FcTransfersApplyCoinAddress struct {
	Id             int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	ApplyId        int64     `json:"apply_id" xorm:"not null comment('出账申请ID') INT(11)"`
	ApplyCoinId    int       `json:"apply_coin_id" xorm:"not null comment('出账申请币种的编号，对应apply_coin表中的id') INT(11)"`
	Address        string    `json:"address" xorm:"not null comment('单条地址，或者all表示全部') VARCHAR(255)"`
	AddressFlag    string    `json:"address_flag" xorm:"not null comment('出账地址,找零地址,接收地址') ENUM('change','from','to')"`
	Status         int       `json:"status" xorm:"not null default 0 comment('状态,0正常') TINYINT(3)"`
	Lastmodify     time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	ToAmount       string    `json:"to_amount" xorm:"not null default 0.00000000000000000000 comment('接收金额') DECIMAL(50,20)"`
	BanFromAddress string    `json:"ban_from_address" xorm:"not null default '' comment('不可使用此地址出账') VARCHAR(255)"`
}

func (o FcTransfersApplyCoinAddress) Find(cond builder.Cond) ([]*FcTransfersApplyCoinAddress, error) {
	res := make([]*FcTransfersApplyCoinAddress, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
