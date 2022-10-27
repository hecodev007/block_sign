package main

import (
	"domsync/common/conf"
	"domsync/routers"
	"domsync/services"
	"domsync/services/registor"
	"fmt"
	"github.com/walletam/rabbitmq"
	"log"
	_ "net/http/pprof"
)

func main() {
	//go http.ListenAndServe("0.0.0.0:6060", nil)
	cfg := conf.Cfg
	log.SetFlags(log.LstdFlags | log.Llongfile)
	//创建地址观察者
	watcher := services.NewWatchControl(conf.Cfg.Sync.Name, conf.Cfg.Sync.AddressRecover, conf.Cfg.Sync.ContractRecover)
	newScanner, ok := registor.ScanFuncMap[conf.Cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("不支持的币种 %s", cfg.Sync.Name))
	}

	newProcess, ok := registor.ProcFuncMap[conf.Cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", cfg.Sync.Name))
	}
	//创建消息推送服务
	pusher := services.NewPushServer(cfg.Push, watcher)
	//创建区块处理服务
	processor := newProcess(*cfg, cfg.Nodes[cfg.Sync.Name], watcher)
	procserver := services.NewProcServer(processor, 10).SetPusher(pusher)
	//创建区块扫描服务
	scanner := newScanner(*cfg, cfg.Nodes[cfg.Sync.Name], watcher)
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
	go mq.Consume(conf.Cfg.Sync.Name+"_addr", watcher.InsertAddr)
	go mq.Consume(conf.Cfg.Sync.Name+"_contract", watcher.InsertContract)
	//注册handler
	router := routers.InitRouter(cfg.Sync.Name, cfg.Mode, *watcher, processor)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Printf(err.Error())
	}
}
