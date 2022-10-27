package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"zcashDataServer/common"
	"zcashDataServer/common/log"
	"zcashDataServer/conf"
	"zcashDataServer/db"
	"zcashDataServer/routers"
	"zcashDataServer/services"
	"zcashDataServer/services/telos"
	"zcashDataServer/services/zec"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var scanFuncMap map[string]ScanFunc
var procFuncMap map[string]ProcFunc

func init() {
	scanFuncMap = map[string]ScanFunc{

		"zec":   zec.NewScanner,
		"telos": telos.NewScanner,
	}
	procFuncMap = map[string]ProcFunc{

		"zec":   zec.NewProcessor,
		"telos": telos.NewProcessor,
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

	//初始化log
	log.InitLogger(cfg.Mode, cfg.Log.Level, cfg.Log.Formatter, cfg.Log.OutFile, cfg.Log.ErrFile)

	//链接区块数据库
	err = db.InitSyncDB(cfg.DataBases["sync"])
	if err != nil {
		log.Panic("init sync db err %v", err)
	}
	log.Debug("ConnectDB sync db success ")
	//链接用户地址数据库
	err = db.InitUserDB(cfg.DataBases["user"])
	if err != nil {
		log.Panic("init user db err %v", err)
	}
	log.Debug("ConnectDB user db success")
	//创建地址观察者
	watcher, err := services.NewWatchControl(cfg.Sync.Name)
	if err != nil {
		log.Panic("new watch control err %v", err)
	}

	log.Debug("NewWatchControl success ")
	newScanner, ok := scanFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("1.don't supported coin %s", cfg.Sync.Name))
	}
	newProcess, ok := procFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", cfg.Sync.Name))
	}
	//创建区块扫描服务
	scanserver, err := services.NewScanServer(newScanner(cfg, cfg.Nodes[cfg.Sync.Name]), cfg)
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
	//注册handler
	r, _, err := routers.InitRouter(cfg.Sync.Name, cfg.Mode, *watcher, processor)
	if err != nil {
		log.Debug("%v", err)
		panic(err)
	}
	log.Infof("startServer  ")
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
		log.Infof("get signal: %v . sever will stop and showdown !", sig)
		if cfg.Sync.EnableSync {
			scanserver.Stop()
		}
		procserver.Stop()
		pusher.Stop()
		if err := s.Shutdown(nil); err != nil {
			log.Fatal("Shutdown server:", err)
		}
		log.Infof("server showdown !")
	}()
	//开启服务
	err = s.ListenAndServe()
	if err != nil {
		log.Error(err.Error())
	}
	log.Info("server end!!!")
}
