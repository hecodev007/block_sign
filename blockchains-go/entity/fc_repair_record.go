package entity

import "github.com/group-coldwallet/blockchains-go/db"

type FcRepairRecord struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Chain      string `json:"chain" xorm:"comment('主链') index VARCHAR(16)"`
	CoinCode   string `json:"chain" xorm:"comment('币种') index VARCHAR(24)"`
	TxId       string `json:"tx_id" xorm:"comment('交易哈希') index VARCHAR(256)"`
	Height     int64  `json:"height" xorm:"comment('高度') BIGINT(20)"`
	Creator    string `json:"creator" xorm:"comment('操作者') index VARCHAR(18)"`
	Remark     string `json:"remark" xorm:"comment('备注') index VARCHAR(128)"`
	CreateTime int64  `json:"create_time" xorm:"comment('创建时间') BIGINT(20)"`
}

func (fr *FcRepairRecord) Insert() error {
	_, err := db.Conn.Insert(fr)
	return err
}
