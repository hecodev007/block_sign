package main

import (
	"flag"
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/job"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

type collectorFunc func(cfg conf.Collect2) cron.Job

var collectorFuncMap = map[string]collectorFunc{
	//"eth": job.NewColdToFeeEthJob,
	"fil": job.NewFixFilAddrJob,

	//"ong": job.NewColdToFeeOngJob,
}

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

	if routineNum <= 0 || routineNum > 10 {
		routineNum = 4
	}
	runtime.GOMAXPROCS(runtime.NumCPU() * routineNum)
	//日志
	util.InitLog()
	if err := conf.LoadConfig3(cfgFile, &cfg); err != nil {
		panic(err)
	}
	conf.DecryptCfg(&cfg)
	//数据库加载
	db.InitOrm2(cfg.DB)

	//开启定时任务
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	for k, v := range cfg.Collectors {
		collectfunc, ok := collectorFuncMap[k]
		if !ok {
			panic(fmt.Errorf("don't find any collector %s", k))
		}

		jobrunner.Schedule(v.Spec, collectfunc(*v))
	}

	//starCronJob()// 添加定时归集任务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("get signal: %v. sever will stop and showdown !", sig)

	jobrunner.Stop()
	log.Println("server showdown !")
}
