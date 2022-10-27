package entity

type FcEosBlock struct {
	Id          int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	Coin        string `json:"coin" xorm:"not null VARCHAR(100)"`
	BlockHeight int    `json:"block_height" xorm:"default 0 index INT(11)"`
	Hash        string `json:"hash" xorm:"comment('块hash') index VARCHAR(150)"`
}
