package db

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"time"
	"xorm.io/xorm"
)

var ConnAddrMgr *xorm.Engine

func InitXOrmAddrMgr() {
	var (
		err    error
		master *xorm.Engine
	)

	if conf.Cfg.DBAddrMgr.Master == "" {
		return
	}
	//主库添加
	master, err = xorm.NewEngine("mysql", conf.Cfg.DBAddrMgr.Master)
	if err != nil {
		log.Fatalf("DB addrMgr error：%s", err.Error())
	}

	master.SetMaxIdleConns(10)
	master.SetMaxOpenConns(200)
	master.SetConnMaxLifetime(60 * time.Second)
	if conf.Cfg.Env == "debug" {
		master.ShowSQL(true)
	}
	master.ShowExecTime(true)
	if err != nil {
		log.Fatalf("init addrMgr master data server error:%s", err.Error())
	}
	ConnAddrMgr = master
}
