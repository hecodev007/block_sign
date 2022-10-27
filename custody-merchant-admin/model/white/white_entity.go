package white

import "time"

type WhiteList struct {
	State       int       `json:"state" gorm:"column:state"`
	Use         int       `json:"use" gorm:"column:use"`
	Id          int64     `json:"id" gorm:"column:id; PRIMARY_KEY"`
	CoinId      int       `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ChainId     int       `gorm:"column:chain_id" json:"chain_id,omitempty"`
	ServiceId   int       `gorm:"column:service_id" json:"service_id,omitempty"`
	CoinName    string    `gorm:"column:coin_name" json:"coin_name,omitempty"`
	ChainName   string    `gorm:"column:chain_name" json:"chain_name,omitempty"`
	ServiceName string    `gorm:"column:service_name" json:"service_name,omitempty"`
	Address     string    `json:"address" gorm:"column:address"`
	AddressName string    `json:"address_name" gorm:"column:address_name"`
	Remark      string    `json:"remark" gorm:"column:remark"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time"`
}

func (w *WhiteList) TableName() string {
	return "white_list"
}
