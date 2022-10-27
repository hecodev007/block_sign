package dao

import (
	"github.com/group-coldwallet/blockchains-go/entity"
	"xorm.io/xorm"
)


func InsertNewMchMoneyItem(db *xorm.Session, item entity.FcMchMoney) (err error) {
	_, err = db.Insert(item)
	return  err
}