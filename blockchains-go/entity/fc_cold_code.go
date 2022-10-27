package entity

type FcColdCode struct {
	CoinId   int `json:"coin_id" xorm:"unique INT(11)"`
	ClodCode int `json:"clod_code" xorm:"default 0 INT(11)"`
}
