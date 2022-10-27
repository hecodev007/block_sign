package main

import (
	"context"
	"github.com/bamzi/jobrunner"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/middleware"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router"
	"github.com/group-coldwallet/blockchains-go/runtime"
	"github.com/group-coldwallet/blockchains-go/runtime/job"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Info("Cold Server ...")
	//
	//if reply, err := redisHelper.SetNx(interceptKey); err == nil && reply == 1 {
	//	redisHelper.Expire(interceptKey, 86400) //60秒过期
	//} else {
	//	log.Println("123")
	//}
	//配置文件
	conf.InitConfig()
	//数据库加载
	db.InitOrm()
	db.InitXOrmAddrMgr()
	//记载第二数据库
	db.InitXOrm2()
	//配置redis连接池
	util.CreateRedisPool(conf.Cfg.Redis.Url, conf.Cfg.Redis.User, conf.Cfg.Redis.Password)
	//加载全局参数
	runtime.InitGlobal()
	//加载IM工具通知
	runtime.InitIM()
	runtime.InitDingRole(conf.Cfg.Env)
	// redis集群
	redis.InitRedis(conf.Cfg.ClusterRedisConfig.Addr, conf.Cfg.ClusterRedisConfig.Pwd, conf.Cfg.ClusterRedisConfig.Cluster)
	redis.InitRedis2(conf.Cfg.ClusterRedisConfig2.Addr, conf.Cfg.ClusterRedisConfig2.Pwd, conf.Cfg.ClusterRedisConfig2.Cluster)
	//启动服务器
	startup()
}

func startup() {
	if conf.Cfg.Env == gin.ReleaseMode {
		//生产模式
		gin.SetMode(gin.ReleaseMode)
	} else {
		//开发模式
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()

	//定时任务开启
	jobRunner()

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	tm := job.NewTxManager()
	tm.StartAsync(cancelCtx)

	loadruoter(r, tm)

	ginpprof.Wrap(r)
	srv := &http.Server{
		Addr:         ":" + conf.Cfg.Http.Port,
		ReadTimeout:  conf.Cfg.Http.ReadTimeout * time.Second,
		WriteTimeout: conf.Cfg.Http.WriteTimeout * time.Second,
		Handler:      r,
	}
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit

	cancelFunc()

	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")

}

func loadruoter(r *gin.Engine, tm *job.TxManager) {

	//跨域请求设置
	r.Use(middleware.GinCors())

	//开发模式打印请求body参数
	if gin.Mode() == gin.DebugMode {
		r.Use(middleware.GinPrintParams())
	}

	//加载路由
	router.InitAdminRouter(r)

	//公共路由
	router.InitCommonRouter(r, tm)

	//测试API
	router.InitTestRouter(r)

	//v1版本API
	router.InitV1Router(r)
	router.InitV3Router(r)
	router.InitCustodyRouter(r) //托管管理后台
	//router.InitV2(r)
	//router.InitV3(r)
	//...

}

//定时任务开启
func jobRunner() {
	log.Info("定时任务开启")
	jobrunner.Start()
	jobrunner.Schedule("@every 240s", job.BalanceJob{Second: 250})
	jobrunner.Schedule("@every 10s", job.TransferApplyCallBackJob{})
	jobrunner.Schedule("@every 30s", job.TransferApplyJob{})

	jobrunner.Schedule("@every 20s", job.TransferApplyBaseJob{CoinName: "heco", LimitNum: 15, SleepSecond: 1})
	jobrunner.Schedule("@every 20s", job.TransferApplyBaseJob{CoinName: "trx", LimitNum: 15, SleepSecond: 1})
	jobrunner.Schedule("@every 20s", job.TransferApplyBaseJob{CoinName: "eth", LimitNum: 15, SleepSecond: 1})

	jobrunner.Schedule("@every 20s", job.TransferApplyBaseJob{CoinName: "hsc", LimitNum: 15, SleepSecond: 1})
	jobrunner.Schedule("@every 20s", job.TransferApplyBaseJob{CoinName: "bsc", LimitNum: 15, SleepSecond: 1})
	//归集任务
	//jobrunner.Schedule("@every 180s", job.CollectZvcJob{}) //归集任务60s测试
	//jobrunner.Schedule("@every 300s", job.NewCollectKlayJob())
}

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o=blockchains-go
