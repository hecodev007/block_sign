package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/zecserver/conf"
	"github.com/group-coldwallet/zecserver/router"
	"github.com/group-coldwallet/zecserver/service/loadkeyservice"
	"github.com/group-coldwallet/zecserver/util"
	"github.com/group-coldwallet/zecserver/util/rylink"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o=zecserver_linux
//CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o=zecserver_windows
//go build -ldflags "-s -w" -o=zecserver_mac
func main() {

	gin.SetMode(conf.GlobalConf.RunModel)

	if conf.GlobalConf.EnableRPC {
		//初始化rpc看客户端
		rylink.ZecRpcClient = rylink.NewZecClient(&util.RpcConnConfig{
			Host: conf.GlobalConf.ZecCfg.RpcHost,
			User: conf.GlobalConf.ZecCfg.RpcUser,
			Pass: conf.GlobalConf.ZecCfg.RpcPassword,
		})

	}

	//前置任务
	//loadTask(conf.GlobalConf)

	r := gin.New()

	//Recovery 中间件会恢复(recovers) 任何恐慌(panics) 如果存在恐慌，中间件将会写入500
	r.Use(gin.Recovery())

	//路由策略
	router.InitRouter(conf.GlobalConf, r)

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

	//加载内存地址文件 Users/zwj/gopath/src/createaddr/files/zec
	var loadkeyService loadkeyservice.BasicService = new(loadkeyservice.LoadService)

	//loadkeyService.ReadFile("/Users/zwj/gopath/src/createaddr/files/btc/test/btc_a_usb_new.csv", "/Users/zwj/gopath/src/createaddr/files/btc/test/btc_b_usb_new.csv")

	//加载历史遗留旧文件，地址下标在1
	//loadkeyService.ReadOleFolder(cfg.ZecCfg.AddrPath)

	//加载本地新文件目录,地址在前
	loadkeyService.ReadNewFolder(cfg.ZecCfg.AddrPath)

}
