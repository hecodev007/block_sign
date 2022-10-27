package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/builder"
)

type CheckApply struct {
	Id       int64  `json:"id,omitempty" gorm:"column:id"`
	ApplyId  int64  `json:"apply_id" gorm:"column:apply_id"`
	Content  string `json:"content,omitempty" gorm:"column:content"`     //加密内容
	CreateAt int64  `json:"create_at,omitempty" gorm:"column:create_at"` //时间戳
	UpdateAt int64  `json:"update_at,omitempty" gorm:"column:update_at"`
}

func (o *CheckApply) Add() (int64, error) {
	return db.Conn2.InsertOne(o)
}

func (o *CheckApply) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
