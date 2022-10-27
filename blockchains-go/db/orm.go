package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
)

//数据库引擎组，读写分离,读库使用轮询访问负载策略.
//写操作和数据库事务操作都将在Master数据库中执行。而读操作则是依据负载策略在某个Slave中执行。
//对外使用dbEngineGroup
var (
	Conn *xorm.EngineGroup
)

func InitOrm() {
	var err error
	Conn, err = LoadOrm(conf.Cfg.DB.Master, conf.Cfg.DB.Slaves, conf.Cfg.Env)
	if err != nil {
		panic(err)
	}

}
func InitOrm2(cfg conf.DBConfig) {
	var err error
	Conn, err = LoadOrm(cfg.Master, cfg.Slaves, "")
	//log.Infof("mastter:%s", cfg.Master)
	//log.Infof("slaves:%v", cfg.Slaves)
	if err != nil {
		panic(err)
	}
}

func LoadOrm(masterUrl string, slaveArr []string, env string) (*xorm.EngineGroup, error) {
	var (
		err    error
		master *xorm.Engine
		slaves []*xorm.Engine
		conn   *xorm.EngineGroup
	)
	//主库添加
	master, err = xorm.NewEngine("mysql", masterUrl)

	if err != nil {
		log.Fatalf("DB error：%s", err.Error())
	}
	master.SetMaxIdleConns(10)
	master.SetMaxOpenConns(200)
	master.SetConnMaxLifetime(60 * time.Second)
	if env == "debug" {
		master.ShowSQL(true)
	}
	master.ShowExecTime(true)
	if err != nil {
		log.Fatalf("init master dataserver error:%s", err.Error())
		return nil, err
	}
	//从库添加
	for _, s_dsn := range slaveArr {
		_dbs, err := xorm.NewEngine("mysql", s_dsn)
		_dbs.SetMaxIdleConns(10)
		_dbs.SetMaxOpenConns(200)
		//if env == "debug" {
		//	_dbs.ShowSQL(true)
		//}
		_dbs.ShowSQL(true)
		_dbs.ShowExecTime(true)
		if err != nil {
			log.Errorf("slave:%s,error:%s", s_dsn, err.Error())
			return nil, err
		} else {
			slaves = append(slaves, _dbs)
		}
	}

	//如果是多个从库，使用轮询访问负载策略，xorm.RoundRobinPolicy()
	//如需权重，随机策略 请查看文档
	conn, err = xorm.NewEngineGroup(master, slaves, xorm.RoundRobinPolicy())
	if err != nil {
		log.Fatalf("init dataserver error:%s", err.Error())
		return nil, err
	}
	//conn.ShowSQL(true)
	conn.SetTableMapper(core.SnakeMapper{})
	conn.SetColumnMapper(core.SnakeMapper{})
	return conn, nil
}
