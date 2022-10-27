package main

import (
	"fmt"
	"github.com/walletam/rabbitmq"
	"golang.org/x/net/context"
	"hecosync/common/conf"
	"hecosync/common/log"
	"hecosync/routers"
	"hecosync/services"
	eth "hecosync/services/wtc"
	"hecosync/utils/dingding"
	_ "net/http/pprof"
)

func main() {
	//go http.ListenAndServe("0.0.0.0:6060", nil)
	cfg := conf.Cfg
	dingding.InitDingBot(context.Background())
	//创建地址观察者
	watcher := services.NewWatchControl(conf.Cfg.Sync.CoinName, conf.Cfg.Sync.AddressRecover, conf.Cfg.Sync.ContractRecover)
	//创建消息推送服务
	pusher := services.NewPushServer(cfg.Push, watcher)
	//创建区块处理服务
	processor := eth.NewProcessor(*cfg, cfg.Node, watcher)
	procserver := services.NewProcServer(processor).SetPusher(pusher)
	//创建区块扫描服务
	scanner := eth.NewScanner(*cfg, cfg.Node, watcher)
	scanserver := services.NewScanServer(scanner, *cfg).SetProcessor(procserver)
	//开启推送服务
	pusher.Start()
	//开启数据处理服务
	procserver.Start()
	//开启区块扫描服务
	if cfg.Sync.EnableSync {
		scanserver.Start()
	}
	mq := rabbitmq.NewRabbitMq(cfg.Mq.HostPort, cfg.Mq.Username, cfg.Mq.Password)
	go mq.Consume(conf.Cfg.Sync.CoinName+"_addr", watcher.InsertAddr)
	go mq.Consume(conf.Cfg.Sync.CoinName+"_contract", watcher.InsertContract)
	//注册handler
	router := routers.InitRouter(cfg.Sync.CoinName, *watcher, processor)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Error(err.Error())
	}
}
