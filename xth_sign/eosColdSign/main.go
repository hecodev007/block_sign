package main

import (
	api2 "bosSign/api"
	"bosSign/common/conf"
	"bosSign/common/log"
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
	log.InitLogger(cfg.RunMode, "info", "text", cfg.OutFile, cfg.ErrFile)
	r, err := api2.InitRouter(cfg.RunMode)
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.HttpPort),
		Handler: r,
		//ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		//WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println(s.Addr)
	log.Info("server start success" + s.Addr)

	go s.ListenAndServe()

	gracefulExitWeb(s)

}
func gracefulExitWeb(server *http.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch

	log.Infof("get signal:%d, server stoping...", sig)
	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := server.Shutdown(cxt)
	if err != nil {
		log.Fatal("stop server err:", err)
	}
	// 看看实际退出所耗费的时间
	log.Infof("server stop,cost time:%v\n", time.Since(now))
}
