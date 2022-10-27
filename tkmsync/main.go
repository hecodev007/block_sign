package main

import (
	"flag"
	"fmt"
	"github.com/walletam/rabbitmq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/db"
	"rsksync/routers"
	"rsksync/services"
	"rsksync/services/kardia"
	"runtime"
	"syscall"
	"time"
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var scanFuncMap map[string]ScanFunc
var procFuncMap map[string]ProcFunc

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	scanFuncMap = map[string]ScanFunc{
		"tkm": kardia.NewScanner,
	}
	procFuncMap = map[string]ProcFunc{
		"tkm": kardia.NewProcessor,
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
	//fmt.Printf("LoadConfig : %v", cfg)
	//初始化log
	//log.InitLogger(cfg.Mode, cfg.Log.Level, cfg.Log.Formatter, cfg.Log.OutFile, cfg.Log.Formatter)
	//链接区块数据库
	err = db.InitSyncDB(cfg.DataBases["sync"])
	if err != nil {
		log.Panic("init sync db err %v", err)
	}
	log.Println("ConnectDB sync db success ")
	//链接用户地址数据库
	err = db.InitUserDB(cfg.DataBases["user"])
	if err != nil {
		log.Panic("init user db err %v", err)
	}
	log.Println("ConnectDB user db success")
	//创建地址观察者
	watcher, err := services.NewWatchControl(cfg.Sync.Name, conf.Sync.AddressRecover, conf.Sync.ContractRecover)
	if err != nil {
		log.Panic("new watch control err %v", err)
	}
	//定义一个全局使用的地址库
	services.WatchCtl = watcher

	log.Println("NewWatchControl success ")
	newScanner, ok := scanFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("1.don't supported coin %s", cfg.Sync.Name))
	}
	newProcess, ok := procFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", cfg.Sync.Name))
	}
	//创建区块扫描服务
	scanserver, err := services.NewScanServer(newScanner(cfg, cfg.Nodes[cfg.Sync.Name], watcher), cfg)
	if err != nil {
		log.Panic("new scan server err %v", err)
	}
	processor := newProcess(cfg, cfg.Nodes[cfg.Sync.Name], watcher)
	//创建区块处理服务
	procserver, err := services.NewProcServer(processor, 10)
	if err != nil {
		log.Panic("new proc server err %v", err)
	}
	//创建消息推送服务
	fmt.Println(cfg.Push)
	pusher, err := services.NewPushServer(cfg.Push, watcher)
	if err != nil {
		log.Panic("new push server err %v", err)
	}
	//链接推送服务到处理服务
	procserver.SetPusher(pusher)
	//链接处理服务到扫描服务
	scanserver.SetProcessor(procserver)
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
	r, _, err := routers.InitRouter(cfg.Sync.Name, cfg.Mode, *watcher, processor)
	if err != nil {
		log.Println("%v", err)
		panic(err)
	}
	log.Printf("startServer  ")
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//注册系统信号监听
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("get signal: %v . sever will stop and showdown !", sig)
		if cfg.Sync.EnableSync {
			scanserver.Stop()
		}
		procserver.Stop()
		pusher.Stop()
		if err := s.Shutdown(nil); err != nil {
			log.Fatal("Shutdown server:", err)
		}
		log.Printf("server showdown !")
	}()
	//开启服务
	err = s.ListenAndServe()
	if err != nil {
		log.Println(err.Error())
	}
}
