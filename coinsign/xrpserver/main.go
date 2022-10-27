package main

import (
	"fmt"
	"net/http"
	"xrpserver/api"
	"xrpserver/common/conf"
	"xrpserver/common/log"
)
func main(){
	err := conf.LoadConfig("./app.toml",conf.Cfg)
	if err != nil {
		panic(err.Error())
	}
	log.InitLogger(conf.Cfg.Mode, conf.Cfg.Log.Level, conf.Cfg.Log.Formatter, conf.Cfg.Log.OutFile, conf.Cfg.Log.ErrFile)

	r := api.InitRouter(conf.Cfg.Cointype)
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", conf.Cfg.Server.Port),
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}
	log.Infof("startServer  %s", s.Addr)
	err = s.ListenAndServe()
	if err != nil {
		log.Infof("startServer  %s %s", s.Addr, err.Error())
	}
}