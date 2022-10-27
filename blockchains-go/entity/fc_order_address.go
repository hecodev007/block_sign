package entity

type FcOrderAddress struct {
	Id           int64  `json:"id" xorm:"pk autoincr BIGINT(11)"`
	OrderId      int64  `json:"order_id" xorm:"default 0 comment('申请id') index BIGINT(11)"`
	Dir          int    `json:"dir" xorm:"default 0 comment('0:from, 1: to') TINYINT(1)"`
	Address      string `json:"address" xorm:"default '' comment('地址') VARCHAR(255)"`
	Amount       int64  `json:"amount" xorm:"default 0 comment('金额 ') BIGINT(20)"`
	CreateAt     int64  `json:"create_at" xorm:"default 0 BIGINT(11)"`
	UpdateAt     int64  `json:"update_at" xorm:"default 0 comment('最后修改时间') BIGINT(20)"`
	TxId         string `json:"tx_id" xorm:"VARCHAR(100)"`
	MuxId        string `json:"mux_id" xorm:"comment('btm sign id') VARCHAR(100)"`
	Vout         int    `json:"vout" xorm:"INT(11)"`
	Scriptpubkey string `json:"scriptpubkey" xorm:"VARCHAR(255)"`
	TokenAmount  int64  `json:"token_amount" xorm:"default 0 comment('代币金额') BIGINT(20)"`
	OrderNo      string `json:"order_no" xorm:"default '' comment('交易订单号') VARCHAR(80)"`
}
