package main

import (
	"fmt"
	api2 "iotasign/api"
	"iotasign/common/conf"
	"iotasign/common/log"
	syslog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	syslog.SetFlags(syslog.LstdFlags | syslog.Llongfile)
}
func main() {
	var err error
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v ", err)
		}
	}()
	//flag.StringVar(&cfgFile, "c", "", "set the yaml conf file")
	//	curl --location --request POST 'http://127.0.0.1:14042/v1/iota/createaddr' --header 'Content-Type: application/json' --data-raw '{"mchId":"goapi","orderId":"goapi","num":30,"coinName":"iota"}'

	//初始化log
	cfg := conf.GetConfig()
	//fmt.Println("cfg.Log.OutFile",cfg.Log.OutFile)
	log.InitLogger(cfg.Log.Level, cfg.Mode, cfg.Log.Formatter, cfg.Log.OutFile, cfg.Log.ErrFile)
	if err != nil {
		log.Panic(err)
	}

	r, err := api2.InitRouter(cfg.Name, cfg.Mode)
	if err != nil {
		log.Panic(err)
	}
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Infof("get signal: %d sever will stop and showdown", sig)

		if err := s.Shutdown(nil); err != nil {
			log.Fatal("Shutdown server:", err)
		}
		log.Infof("server showdown !")
	}()
	log.Infof("startServer  %s", s.Addr)
	err = s.ListenAndServe()
	if err != nil {
		log.Infof("startServer  %s, err:%v", s.Addr, err.Error())
	}

}
