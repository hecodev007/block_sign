package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
)

type ReportType int

const (
	ColdAddrBalanceNotEnough ReportType = 1 //没有符合条件的出账地址
	CollectAmountNotEnough   ReportType = 2
	AuditTimeout             ReportType = 10
)

type FcReportRecord struct {
	Id           int        `json:"id" xorm:"not null pk autoincr INT(11)"`
	Chain        string     `json:"chain" xorm:"comment('主链') index VARCHAR(16)"`
	CoinCode     string     `json:"chain" xorm:"comment('币种') index VARCHAR(24)"`
	TxId         string     `json:"tx_id" xorm:"comment('交易哈希')  VARCHAR(256)"`
	ReportType   ReportType `json:"report_type" xorm:"comment('类型')  INT(11)"`
	OuterOrderId string     `json:"outer_order_id" xorm:"comment('订单id') index VARCHAR(256)"`
	Remark       string     `json:"remark" xorm:"comment('备注') LONGTEXT"`
	CreateTime   int64      `json:"create_time" xorm:"comment('创建时间') BIGINT(20)"`
}

func (fr *FcReportRecord) Insert() error {
	_, err := db.Conn.Insert(fr)
	return err
}
