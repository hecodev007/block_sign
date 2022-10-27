package main

/*
套用wallet-sign框架服务，方便以后维护
*/

import (
	"flag"
	"fmt"
	"github.com/eth-sign/conf"
	"github.com/eth-sign/routers"
	v1 "github.com/eth-sign/services/v1"
	"github.com/gin-gonic/gin"
	//log "log"
	"log"
	"strings"
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

	// 设置日志格式为json
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFlags(log.LstdFlags | log.Llongfile)
	// 初始化配置文件
	conf.InitConfig()
	flag.Parse()
	if offline {
		if nums <= 0 {
			log.Printf("generate key numbers is less than zero")
			return
		}
		srv := v1.GetIService()
		// fmt.Println(conf.Config.CoinType,conf.Config.MchId,conf.Config.OrderId)
		err := srv.MultiThreadCreateAddrService(nums, conf.Config.CoinType, conf.Config.MchId, conf.Config.OrderId)
		if err != nil {
			log.Printf("generate key error,Err=[%v]", err)
		}
		return
	}
	log.Printf("start %s wallet sign service", conf.Config.CoinType)
	if !conf.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	path := fmt.Sprintf("%s/%s", strings.ToLower(conf.Config.Version), strings.ToLower(conf.Config.CoinType))
	group := r.Group(path)
	// 初始化路由
	routers.InitRouters(group)
	// 启动
	r.Run(":" + conf.Config.Port)
}
