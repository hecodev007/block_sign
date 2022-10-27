package db

import (
	"algoDataServer/common/conf"
	"algoDataServer/common/log"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func init() {
	if err := InitSyncDB(conf.Cfg.DataBases["sync"]); err != nil {
		panic(err.Error())
	}
	if err := InitUserDB(conf.Cfg.DataBases["user"]); err != nil {
		panic(err.Error())
	}
}

var SyncDB *MysqlDB
var UserDB *MysqlDB

type MysqlDB struct {
	Name  string
	DB    *gorm.DB
	IDGen *snowflake.Node
}

//连接数据库
func InitSyncDB(cfg conf.DatabaseConfig) error {
	dburl := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", cfg.User, cfg.PassWord, cfg.Url, cfg.Name)
	db, err := InitDataBase(cfg.Type, dburl, cfg.Mode)
	if err != nil {
		return err
	}
	node, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}
	SyncDB = &MysqlDB{
		Name:  cfg.Name,
		DB:    db,
		IDGen: node,
	}
	return nil
}

//连接数据库
func InitUserDB(cfg conf.DatabaseConfig) error {
	dburl := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", cfg.User, cfg.PassWord, cfg.Url, cfg.Name)
	db, err := InitDataBase(cfg.Type, dburl, cfg.Mode)
	if err != nil {
		return err
	}
	node, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}
	UserDB = &MysqlDB{
		Name:  cfg.Name,
		DB:    db,
		IDGen: node,
	}
	return nil
}
func InitDataBase(dbType, dbUrl, dbMode string) (*gorm.DB, error) {
	if dbUrl == "" || dbType == "" {
		log.Infof("database's conf is null")
	}
	db, err := gorm.Open(dbType, dbUrl)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("gorm db is nil")
	}
	db.SingularTable(true)
	db.DB().SetMaxIdleConns(2)
	db.DB().SetMaxOpenConns(32)
	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.LogMode(dbMode == "dev")
	return db, nil
}
