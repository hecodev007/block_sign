package entity

type FcAccessSalt struct {
	SaltId    int    `json:"salt_id" xorm:"not null pk autoincr INT(11)"`
	AccessId  int    `json:"access_id" xorm:"not null INT(11)"`
	EcSalt    string `json:"ec_salt" xorm:"not null VARCHAR(5)"`
	TokenSalt string `json:"token_salt" xorm:"not null default '' VARCHAR(5)"`
	Secretkey string `json:"secretkey" xorm:"not null VARCHAR(64)"`
}
