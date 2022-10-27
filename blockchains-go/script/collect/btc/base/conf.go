package base

import (
	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"xorm.io/xorm"
)

var Cfg *conf.Config
var CoinCfg *CoinConf
var Conn *xorm.EngineGroup
var ErrDingBot *dingding.DingBot
var CoinSet *entity.FcCoinSet

func InitConf(baseConfPath, coinPath string) {
	var err error

	//配置文件
	//币种配置
	CoinCfg = &CoinConf{}

	_, err = toml.DecodeFile(coinPath, CoinCfg)
	if err != nil {
		log.Infof("读取错误,err:%s，使用默认配置", err.Error())
		CoinCfg.New()
	} else {
		err = CoinCfg.Check()
		if err != nil {
			panic(err)
		}
	}
	Cfg = conf.LoadConfig()

	Conn, err = db.LoadOrm(Cfg.DB.Master, Cfg.DB.Slaves, CoinCfg.Env)
	if err != nil {
		panic(err)
	}

	//redis
	util.CreateRedisPool(Cfg.Redis.Url, Cfg.Redis.User, Cfg.Redis.Password)

	//钉钉通知
	ErrDingBot = &dingding.DingBot{
		Name:   "ding-robot-merge-btc",
		Token:  "e73c7441c796143b2c374b0f5a87efc59bdd37c950805c831d5cd46014e9814d",
		Source: make(chan []byte, 50),
		Quit:   make(chan struct{}),
	}
	ErrDingBot.Start()

	//初始化币种信息
	CoinSet, err = FindCoin(BtcCoinName)
	if err != nil {
		panic(err.Error())
	}

}
