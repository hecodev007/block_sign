package db

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"time"
	"xorm.io/xorm"
)

var Conn2 *xorm.Engine

func InitXOrm2() {
	var (
		err    error
		master *xorm.Engine
	)

	//主库添加
	master, err = xorm.NewEngine("mysql", conf.Cfg.DB2.Master)
	if err != nil {
		log.Fatalf("DB2 error：%s", err.Error())
	}
	master.SetMaxIdleConns(10)
	master.SetMaxOpenConns(200)
	master.SetConnMaxLifetime(60 * time.Second)
	if conf.Cfg.Env == "debug" {
		master.ShowSQL(true)
	}
	master.ShowExecTime(true)
	if err != nil {
		log.Fatalf("init db2 master data server error:%s", err.Error())
	}
	Conn2 = master
}
