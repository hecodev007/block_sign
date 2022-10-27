package launcher

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/nep5server/conf"
	"github.com/group-coldwallet/nep5server/middleware"
	"github.com/group-coldwallet/nep5server/router"
)

func InitGin(cfg *conf.Config) *gin.Engine {
	//设置gin框架开发模式
	if cfg.Dev == "release" {
		gin.SetMode("release")
	} else {
		gin.SetMode("debug")
	}
	r := gin.New()
	//Recovery 中间件会恢复(recovers) 任何恐慌(panics) 如果存在恐慌，中间件将会写入500
	r.Use(gin.Recovery())
	//跨域请求
	r.Use(middleware.GinCors())
	//全局自定义日志中间件设置,每24小时切割,某些API单独的中间件禁止全局使用,控制台不再打印
	r.Use(middleware.GinLogger(cfg.LogCfg.LogPath, cfg.LogCfg.LogSPath, cfg.LogCfg.LogName, cfg.LogCfg.LogLevel))
	if gin.Mode() == gin.DebugMode {
		//参数打印
		r.Use(middleware.GinPrintParams())
	}
	router.InitV1(r, cfg)

	return r
}
