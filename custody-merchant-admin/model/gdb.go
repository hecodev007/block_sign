package model

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	_ "gorm.io/plugin/dbresolver"
	"regexp"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	. "custody-merchant-admin/config"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/log"
)

var GDB *gorm.DB
var DBCacheStore cache.CacheStore

func DB() *gorm.DB {

	if GDB == nil {
		log.Debugf("Model NewDB")
		newDb, err := NewDB()
		if err != nil {
			panic(err)
		}
		sqlDB, err := newDb.DB()
		if err != nil {
			fmt.Errorf("connect db server failed.")
		}
		sqlDB.SetMaxIdleConns(10)                  // SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxOpenConns(100)                 // SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetConnMaxLifetime(time.Second * 60) // SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		GDB = newDb
	}

	return GDB
}

func NewDB() (*gorm.DB, error) {
	cb := Conf.DB["web"]
	dsn := cb.UserName + ":" + cb.Pwd + "@tcp(" + cb.Host + ":" + cb.Port + ")/" + cb.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
	mySQL := mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	})

	db, err := gorm.Open(mySQL, &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Error),
		//Logger: logger.Default.LogMode(logger.Info),
		Logger:      logger.Default.LogMode(logger.Warn),
		PrepareStmt: true,
	})

	// 主从
	//db.Use(dbresolver.Register(dbresolver.Config{
	//	// `db2` 作为 sources，`db3`、`db4` 作为 replicas
	//	Sources:  []gorm.Dialector{mysql.Open("db2_dsn")},
	//	Replicas: []gorm.Dialector{mysql.Open("db3_dsn"), mysql.Open("db4_dsn")},
	//	// sources/replicas 负载均衡策略
	//	Policy: dbresolver.RandomPolicy{},
	//}).Register(dbresolver.Config{
	//	// `db1` 作为 sources（DB 的默认连接），对于 `User`、`Address` 使用 `db5` 作为 replicas
	//	Replicas: []gorm.Dialector{mysql.Open("db5_dsn")},
	//}, &User{}, &Address{}).Register(dbresolver.Config{
	//	// `db6`、`db7` 作为 sources，对于 `orders`、`Product` 使用 `db8` 作为 replicas
	//	Sources:  []gorm.Dialector{mysql.Open("db6_dsn"), mysql.Open("db7_dsn")},
	//	Replicas: []gorm.Dialector{mysql.Open("db8_dsn")},
	//}, "orders", &Product{}, "secondary"))

	if err != nil {
		return nil, err
	}
	return db, nil
}

func ModelError(db *gorm.DB, msg string) error {
	pc, file, line, _ := runtime.Caller(1)
	rt := runtime.FuncForPC(pc)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warnf("%s,文件:%s:%s:%d, error: %v", msg, file, rt.Name(), line, db.Error)
			return nil
		}
		errInfo := fmt.Sprintf("%s,文件:%s:%s:%d, error: %v", msg, file, rt.Name(), line, err)
		log.Errorf(errInfo)
		return errors.New(err.Error())
	}
	return nil
}

func CacheStore() cache.CacheStore {
	if DBCacheStore == nil {
		switch Conf.CacheStore {
		case REDIS:
			DBCacheStore = cache.NewRedisCache(Conf.Redis.AloneAddress, Conf.Redis.AlonePwd, time.Hour)
		default:
			DBCacheStore = cache.NewInMemoryStore(time.Hour)
		}
	}

	return DBCacheStore
}

// FilteredSQLInject
// 正则过滤sql注入的方法
// 参数 : 要匹配的语句
func FilteredSQLInject(toMatchStr ...string) bool {
	//过滤 ‘
	//ORACLE 注解 --  /**/
	//关键字过滤 update ,delete
	// 正则的字符串, 不能用 " " 因为" "里面的内容会转义
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	for _, s := range toMatchStr {
		bl := re.MatchString(s)
		if bl {
			return true
		}
	}
	return false
}
