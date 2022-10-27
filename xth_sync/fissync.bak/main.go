package main

import (
	"fissync/common/conf"
	"fissync/common/db"
	"fissync/common/log"
	"fissync/routers"
	"fissync/services"
	"fissync/services/registor"
	"fmt"
	_ "net/http/pprof"
)

func main() {
	conf.InitConfig()
	db.Init()
	registor.Init()
	log.InitLogger(true, conf.Cfg.Log.Level, conf.Cfg.Log.Formatter, conf.Cfg.Log.OutFile, conf.Cfg.Log.ErrFile)
	//创建地址观察者
	watcher := services.NewWatchControl(conf.Cfg.Sync.Name)
	newScanner, ok := registor.ScanFuncMap[conf.Cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("1.don't supported coin %s", conf.Cfg.Sync.Name))
	}

	newProcess, ok := registor.ProcFuncMap[conf.Cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", conf.Cfg.Sync.Name))
	}
	//创建消息推送服务
	pusher := services.NewPushServer(conf.Cfg.Push, watcher)
	//创建区块处理服务
	processor := newProcess(*conf.Cfg, conf.Cfg.Nodes[conf.Cfg.Sync.Name], watcher)
	procserver := services.NewProcServer(processor, 10).SetPusher(pusher)
	//创建区块扫描服务
	scanner := newScanner(*conf.Cfg, conf.Cfg.Nodes[conf.Cfg.Sync.Name], watcher)
	scanserver := services.NewScanServer(scanner, *conf.Cfg).SetProcessor(procserver)
	//开启推送服务
	pusher.Start()
	//开启数据处理服务
	procserver.Start()
	//开启区块扫描服务
	if conf.Cfg.Sync.EnableSync {
		scanserver.Start()
	}

	//注册handler
	router := routers.InitRouter(conf.Cfg.Sync.Name, conf.Cfg.Mode, *watcher, processor)
	if err := router.Run(fmt.Sprintf(":%s", conf.Cfg.Server.Port)); err != nil {
		log.Error(err.Error())
	}
}