package main

import (
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	"avaxDataServer/routers"
	"avaxDataServer/services"
	"avaxDataServer/services/registor"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()

	cfg := conf.Cfg
	//创建地址观察者
	watcher, err := services.NewWatchControl(cfg.Sync.Name)
	if err != nil {
		panic(fmt.Errorf("new watch control err %v", err))
	}

	newScanner, ok := registor.ScanFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("1.don't supported coin %s", cfg.Sync.Name))
	}

	newProcess, ok := registor.ProcFuncMap[cfg.Sync.Name]
	if !ok {
		panic(fmt.Errorf("2.don't supported coin %s", cfg.Sync.Name))
	}

	//创建区块扫描服务
	scanserver, err := services.NewScanServer(newScanner(*cfg, cfg.Nodes[cfg.Sync.Name]), *cfg)
	if err != nil {
		log.Panic("new scan server err %v", err)
	}

	processor := newProcess(*cfg, cfg.Nodes[cfg.Sync.Name], watcher)

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
			//scanserver.Stop()
		}
		//procserver.Stop()
		//pusher.Stop()
		if err := s.Shutdown(nil); err != nil {
			log.Fatal("Shutdown server:", err)
		}
		log.Infof("server showdown !")
	}()
	//开启服务
	err = s.ListenAndServe()
	log.Infof("server end:%v", err.Error())
}
