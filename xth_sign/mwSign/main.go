package main

import (
	"context"
	"fmt"
	api2 "mwSign/api"
	"mwSign/common/conf"
	"mwSign/common/log"
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
	log.Infof("startServer  %s", s.Addr)

	go s.ListenAndServe()

	gracefulExitWeb(s)

}
func gracefulExitWeb(server *http.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch

	log.Infof("get signal: %d sever will stop and showdown", sig)
	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := server.Shutdown(cxt)
	if err != nil {
		log.Fatal("shutdown server error:", err)
	}
	// 看看实际退出所耗费的时间
	log.Infof("服务关闭,耗时:%v", time.Since(now))
}
