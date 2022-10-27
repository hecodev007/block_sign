package main

import (
	"context"
	"fmt"
	"github.com/group-coldwallet/nep5server/conf"
	"github.com/group-coldwallet/nep5server/launcher"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o=riskcore
func main() {
	//
	conf.InitNacos()
	//配置文件参数，当做全局变量
	cfg := conf.InitConfig("")
	if cfg.PemPath == "" {
		panic("miss PemPath config")
	}
	//加载密钥
	loadKeys(cfg.PemPath)
	//启动http服务
	statrHttp(cfg)

}

func loadKeys(folderpath string) {
	launcher.LoadKeys(folderpath)
}

func statrHttp(conf *conf.Config) {
	//路由初始化
	routerHandler := launcher.InitGin(conf)
	//服务器设置
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", conf.HttpCfg.Port),
		Handler:      routerHandler,
		ReadTimeout:  conf.HttpCfg.ReadTimeout * time.Second,
		WriteTimeout: conf.HttpCfg.WriteTimeout * time.Second,
	}

	go func() {
		logrus.Infof("listen port:%s", conf.HttpCfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()
	//等待中断信号以5秒的超时时间正常关闭服务器。
	quit := make(chan os.Signal, 1)
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
