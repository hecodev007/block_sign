package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/xorm"
)

func TransPushGet(bean interface{}, query interface{}, args ...interface{}) (bool, error) {
	return db.Conn.SQL(query, args...).Get(bean)
}

// obj must be array pEveryOne := make([]*Userinfo)
func TransPushFind(retobj interface{}, query interface{}, args ...interface{}) error {
	return db.Conn.SQL(query, args...).Find(retobj)
}

func TransPushIterator(bean interface{}, query interface{}, iteratorFunc xorm.IterFunc, args ...interface{}) error {
	return db.Conn.Where(query, args...).Iterate(bean, iteratorFunc)
}

func TransPushInterface(sqlargs ...interface{}) ([]map[string]interface{}, error) {
	return db.Conn.QueryInterface(sqlargs...)
}

func TransPushCount(bean interface{}, query interface{}, args ...interface{}) (int64, error) {
	return db.Conn.Where(query, args...).Count(bean)
}

func TransPushInsert(sqlargs ...interface{}) (int64, error) {
	res, err := db.Conn.Exec(sqlargs...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func TransPushInsertOne(bean interface{}) (int64, error) {
	return db.Conn.InsertOne(bean)
}

func TransPushUpdate(sqlargs ...interface{}) (int64, error) {
	res, err := db.Conn.Exec(sqlargs...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func TransPushGetSession() *xorm.Session {
	return db.Conn.NewSession()
}

func TransPushGetDBEnginGroup() *xorm.EngineGroup {
	return db.Conn
}
