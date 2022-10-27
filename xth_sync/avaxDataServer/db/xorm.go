package db

import (
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"xorm.io/xorm"
)

func init() {
	log.Info("XORM INIT")
	InitSyncDB2(conf.Cfg.DataBases["sync"])
	InitUserDB2(conf.Cfg.DataBases["user"])
	fmt.Println("xorm end")
}

var (
	SyncConn *xorm.Engine
	UserConn *xorm.Engine
)

//连接数据库
func InitSyncDB2(cfg conf.DatabaseConfig) error {
	dburl := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", cfg.User, cfg.PassWord, cfg.Url, cfg.Name)
	conn, err := initDBConn(cfg.Type, dburl, cfg.Mode)
	if err != nil {
		return err
	}
	SyncConn = conn
	return nil
}

//连接数据库
func InitUserDB2(cfg conf.DatabaseConfig) error {
	dburl := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", cfg.User, cfg.PassWord, cfg.Url, cfg.Name)
	conn, err := initDBConn(cfg.Type, dburl, cfg.Mode)
	if err != nil {
		return err
	}
	UserConn = conn
	return nil
}
func initDBConn(dbType, dbUrl, dbMode string) (*xorm.Engine, error) {
	if dbUrl == "" || dbType == "" {
		log.Error("database's conf is null")
		panic("")
	}
	conn, err := xorm.NewEngine(dbType, dbUrl)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, fmt.Errorf("gorm db is nil")
	}
	conn.SetMaxIdleConns(2)
	conn.SetMaxOpenConns(6)
	conn.SetConnMaxLifetime(60 * time.Second)
	//conn.ShowSQL(true)
	conn.ShowExecTime(true)
	return conn, nil
}
