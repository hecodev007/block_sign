package main

import (
	api2 "glmrsign/api"
	"glmrsign/common/conf"
	"glmrsign/common/log"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//初始化log
	cfg := conf.GetConfig()
	log.InitLogger(cfg.Log.Level, cfg.Mode, cfg.Log.Formatter, cfg.Log.OutFile, cfg.Log.ErrFile)

	r, err := api2.InitRouter(cfg.Name, cfg.Mode)
	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go gracefulExitWeb(s)

	log.Infof("服务启动成功 %s", s.Addr)
	if err :=s.ListenAndServe();err != nil {
		panic(err.Error())
	}


}
func gracefulExitWeb(server *http.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch

	log.Infof("收到信号:%d, 服务即将停止...", sig)
	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := server.Shutdown(cxt)
	if err != nil {
		log.Fatal("关闭服务错误:", err)
	}
	// 看看实际退出所耗费的时间
	log.Infof("服务关闭,耗时:%v\n", time.Since(now))
}
