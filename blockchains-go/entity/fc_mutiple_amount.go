package entity

type FcMutipleAmount struct {
	Id          int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId      int    `json:"coin_id" xorm:"index(address_id) INT(11)"`
	AddressId   int    `json:"address_id" xorm:"index(address_id) INT(11)"`
	Freeamount  string `json:"freeamount" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Freezamount string `json:"freezamount" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
}
