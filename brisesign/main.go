package main

/*
套用wallet-sign框架服务，方便以后维护
*/

import (
	"brisesign/conf"
	"brisesign/redis"
	"brisesign/services"
	"brisesign/util/dingding"
	"brisesign/websrv"
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	offline bool
	nums    int
)

func init() {
	flag.BoolVar(&offline, "o", false, "this server is offline generate key,default is [false]")
	flag.IntVar(&nums, "n", 1, "generate key numbers,default is [0]")
}
func main() {
	// 初始化配置文件
	conf.InitConfig()

	flag.Parse()

	//log.InitLogger(conf.Config.Debug, conf.Config.Log.Level, conf.Config.Log.Formatter, conf.Config.Log.OutFile, conf.Config.Log.ErrFile)
	log.SetFlags(log.LstdFlags | log.Llongfile)
	if offline {
		doOfflineWork()
		return
	}

	log.Printf("start %s wallet sign service", conf.Config.CoinType)
	if !conf.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	// 初始化Redis
	redis.InitRedis(conf.Config.Redis.Addr, conf.Config.Redis.Pwd)

	// 初始化钉钉机器人
	dingding.InitDingBot(cancelCtx)

	// 初始化并启动HTTP服务
	httpSrv := websrv.NewWebSrv(cancelCtx)
	httpSrv.StartAsync()

	// 等待停止信号
	quitCh := make(chan os.Signal)
	signal.Notify(quitCh, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	<-quitCh

	httpSrv.Stop()
	cancelFunc()
	time.Sleep(3 * time.Second)

	log.Println("签名服务已退出")
}

func doOfflineWork() {
	if nums <= 0 {
		log.Println("generate key numbers is less than zero")
		return
	}
	srv := services.GetIService(nil)
	err := srv.MultiThreadCreateAddrService(nums, conf.Config.CoinType, conf.Config.MchId, conf.Config.OrderId)
	if err != nil {
		log.Println("generate key error,Err=[%v]", err)
	}
}
