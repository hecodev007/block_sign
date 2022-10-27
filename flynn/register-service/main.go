package main

import (
	"flag"
	"fmt"
	"github.com/group-coldwallet/flynn/register-service/db"
	"github.com/group-coldwallet/flynn/register-service/routers"

	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/flynn/register-service/conf"
	log "github.com/sirupsen/logrus"
	"strings"
)

var (
	cfgPath string
)

func init() {
	flag.StringVar(&cfgPath, "f", "./conf/app.toml", "配置文件路径")
}
func main() {
	flag.Parse()
	// 设置日志格式为json
	log.SetFormatter(&log.JSONFormatter{})
	// 初始化配置文件
	conf.InitConfig(cfgPath)
	if !conf.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	//链接用户地址数据库
	err := db.InitUserDB(conf.Config.DataBases["user"])
	if err != nil {
		log.Panic("init user db err %v", err)
	}
	log.Info("ConnectDB user db success")

	r := gin.Default()
	path := fmt.Sprintf("%s", strings.ToLower(conf.Config.Version))
	group := r.Group(path)
	// 初始化路由
	routers.InitRouters(group)
	// 启动
	r.Run(":" + conf.Config.Port)
}
