package base

import "time"

type DictList struct {
	Id         int       `json:"id" gorm:"column:id; PRIMARY_KEY"`
	DictId     int       `json:"dict_id" gorm:"column:dict_id"`
	DictName   string    `json:"dict_name" gorm:"column:dict_name"`
	DictValue  int       `json:"dict_value" gorm:"column:dict_value"`
	Remark     string    `json:"remark" gorm:"column:remark"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (u *DictList) TableName() string {
	return "dict_list"
}
