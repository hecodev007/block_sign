package dao

import (
	"github.com/group-coldwallet/blockchains-go/entity"
	"xorm.io/xorm"
)


func InsertNewMchProfileItem(db *xorm.Session, item entity.FcMchProfile) (err error) {
	_, err = db.Insert(item)
	return  err
}