package main

import (
	"flag"
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		//	err        error
		routineNum int
		cfgFile    string
		cfg        conf.CollectConfig
	)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()

	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "conf/app.toml", "set the toml config file")
	flag.Parse()

	if err := conf.LoadConfig3(cfgFile, &cfg); err != nil {
		panic(err)
	}
	conf.DecryptCfg(&cfg)
	// 数据库加载
	db.InitOrm2(cfg.DB)
	// starCronJob()// 添加定时归集任务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("get signal: %v. sever will stop and showdown !", sig)

	jobrunner.Stop()
	log.Println("server showdown !")
}
