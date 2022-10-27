package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/group-coldwallet/flynn/register-service/conf"
	log "github.com/sirupsen/logrus"

	"time"
	"xorm.io/xorm"
)

var (
	UserConn *xorm.Engine
)

//连接数据库
func InitUserDB(cfg conf.DatabaseConfig) error {
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
		log.Infof("database's conf is null")
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
	conn.ShowSQL(true)
	return conn, nil
}
