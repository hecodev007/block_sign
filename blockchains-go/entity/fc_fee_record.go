package entity

type FcFeeRecord struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address    string `json:"address" xorm:"comment('打手续费地址') VARCHAR(255)"`
	OutOrderid string `json:"out_orderid" xorm:"comment('外部订单号') VARCHAR(100)"`
	AddTime    int    `json:"add_time" xorm:"not null default 0 comment('插入时间') INT(10)"`
}
