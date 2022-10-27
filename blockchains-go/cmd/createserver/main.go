package main

import (
	"flag"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/create"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	var (
		//	err        error
		routineNum int
		cfgFile    string
		cfg        *conf.Config
		svr        *service.CreateService
	)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("process exit err : %v \n", err)
		}
	}()

	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "", "set the yaml config file")
	flag.Parse()

	if routineNum <= 0 || routineNum > 10 {
		routineNum = 4
	}
	runtime.GOMAXPROCS(runtime.NumCPU() * routineNum)

	cfg = conf.LoadConfig(cfgFile)

	svr = service.NewCreateService([]service.Creator{
		create.NewEthCreator(cfg.Creates["eth"].Url),
	})

	//starCronJob()// 添加定时归集任务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		log.Println("get signal: ", sig, ". sever will stop and showdown !")
		svr.Stop()
		log.Println("server showdown !")
	}()

	svr.Start()
}
