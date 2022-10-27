package entity

type FcHooSmallAddress struct {
	Id       int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinName string `json:"coin_name" xorm:"VARCHAR(15)"`
	Address  string `json:"address" xorm:"VARCHAR(80)"`
}
