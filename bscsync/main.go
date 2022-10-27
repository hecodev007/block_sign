package main

import (
	"context"
	"dataserver/common"
	"dataserver/conf"
	"dataserver/db"
	"dataserver/log"
	"dataserver/routers"
	"dataserver/services"
	"dataserver/services/chain"
	"dataserver/utils"
	"dataserver/utils/dingding"
	"flag"
	"fmt"
	"github.com/walletam/rabbitmq"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var scanFuncMap map[string]ScanFunc
var procFuncMap map[string]ProcFunc

func init() {
	scanFuncMap = map[string]ScanFunc{
		"bsc": chain.NewScanner,
	}
	procFuncMap = map[string]ProcFunc{
		"bsc": chain.NewProcessor,
	}
}
func main() {
	var (
		routineNum int
		cfgFile    string
		cfg        conf.Config
	)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()
	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "", "set the toml conf file")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	err := conf.LoadConfig(cfgFile, &cfg)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("LoadConfig : %v", cfg)
	// 初始化log
	// 链接区块数据库
	err = db.InitSyncDB(cfg.DataBases["sync"])
	if err != nil {
		log.Errorf("init sync db err %v", err)
	}
	log.Info("ConnectDB sync db success ")
	// 链接用户地址数据库
	err = db.InitUserDB(cfg.DataBases["user"])
	if err != nil {
		log.Errorf("init user db err %v", err)
	}
	log.Info("ConnectDB user db success")

	dingding.InitDingBot(context.Background())
	dingding.NotifyError("BSC数据服务启动")
	// 创建地址观察者
	watcher, err := services.NewWatchControl(cfg.Sync.Name)
	if err != nil {
		log.Errorf("new watch control err %v", err)
	}
	log.Info("NewWatchControl success ")

	//定时重载地址
	t := utils.NewMyTicker(120, watcher.ReloadWatchAddress)
	defer t.Stop()
	go func() {
		t.Start()
	}()

	newScanner, ok := scanFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("1.don't supported coin %s", cfg.Sync.Name))
	}
	newProcess, ok := procFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", cfg.Sync.Name))
	}
	// 创建区块扫描服务
	scanServer, err := services.NewScanServer(newScanner(cfg, cfg.Nodes[cfg.Sync.Name], watcher), cfg, cfg.Nodes[cfg.Sync.Name].Url)
	if err != nil {
		log.Errorf("new scan server err %v", err)
	}
	processor := newProcess(cfg, cfg.Nodes[cfg.Sync.Name], watcher)
	// 创建区块处理服务
	procServer, err := services.NewProcServer(processor, 10)
	if err != nil {
		log.Errorf("new proc server err %v", err)
	}
	// 创建消息推送服务
	fmt.Println(cfg.Push)
	pusher, err := services.NewPushServer(cfg.Push, watcher)
	if err != nil {
		log.Errorf("new push server err %v", err)
	}
	// 链接推送服务到处理服务
	procServer.SetPusher(pusher)
	// 链接处理服务到扫描服务
	scanServer.SetProcessor(procServer)
	// 开启推送服务
	pusher.Start()

	// 开启数据处理服务
	procServer.Start()
	// 开启区块扫描服务
	if cfg.Sync.EnableSync {
		scanServer.Start()
	}
	mq := rabbitmq.NewRabbitMq(cfg.Mq.HostPort, cfg.Mq.Username, cfg.Mq.Password)
	go mq.Consume(cfg.Sync.Name+"_addr", watcher.InsertAddr)
	go mq.Consume(cfg.Sync.Name+"_contract", watcher.InsertContract)
	// 注册handler
	r, _, err := routers.InitRouter(cfg.Sync.Name, cfg.Mode, *watcher, processor)
	if err != nil {
		log.Infof("%v", err)
		panic(err)
	}
	log.Info("startServer  ")
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// 注册系统信号监听
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Infof("get signal: %v . sever will stop and showdown !", sig)
		if cfg.Sync.EnableSync {
			scanServer.Stop()
		}
		procServer.Stop()
		pusher.Stop()
		if err := s.Shutdown(nil); err != nil {
			log.Fatal("Shutdown server:", err)
		}
		log.Info("server showdown !")
	}()
	// 开启服务
	if err = s.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
