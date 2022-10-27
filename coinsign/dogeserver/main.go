package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/dogeserver/api"
	"github.com/group-coldwallet/dogeserver/conf"
	_ "github.com/group-coldwallet/dogeserver/docs"
	"github.com/group-coldwallet/dogeserver/middleware/ginmiddleware"
	"github.com/group-coldwallet/dogeserver/router"
	"github.com/group-coldwallet/dogeserver/service"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {

	conf.InitConfig()
	api.InitApiService()
	gin.SetMode(conf.GlobalConf.RunModel)

	//前置任务
	loadTask(conf.GlobalConf)

	r := gin.Default()
	//全局跨域中间件设置，某些API单独的中间件禁止全局使用
	r.Use(ginmiddleware.GinCors())
	//全局自定义日志中间件设置,每24小时切割,某些API单独的中间件禁止全局使用,控制台不再打印
	r.Use(ginmiddleware.GinLogger(conf.GlobalConf.LogCfg.LogPath, conf.GlobalConf.LogCfg.LogName, conf.GlobalConf.LogCfg.LogLevel))
	//swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//路由策略
	router.InitV1Router(r)

	//服务器设置
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.GlobalConf.HttpCfg.Port),
		Handler:      r,
		ReadTimeout:  conf.GlobalConf.HttpCfg.ReadTimeout * time.Second,
		WriteTimeout: conf.GlobalConf.HttpCfg.WriteTimeout * time.Second,
	}
	go func() {
		logrus.Infof("listen port:%d", conf.GlobalConf.HttpCfg.Port)
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()
	//等待中断信号以5秒的超时时间正常关闭服务器。
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("Server Shutdown:", err)
	}
	logrus.Info("Server exiting")
}

//启动前任务
func loadTask(cfg *conf.GlobalConfig) {
	loadkeyService := &service.LoadService{}
	//加载内存地址文件 Users/zwj/gopath/src/createaddr/files/btc
	loadkeyService.ReadNewFolder(cfg.DogeCfg.CreateAddrPath)

}
