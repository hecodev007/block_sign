package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/btcsign/api"
	"github.com/group-coldwallet/btcsign/api/v1"
	"github.com/group-coldwallet/btcsign/conf"
	_ "github.com/group-coldwallet/btcsign/docs"
	"github.com/group-coldwallet/btcsign/middleware/ginmiddleware"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func InitRouter(cfg *conf.GlobalConfig, r *gin.Engine) {

	//===================================================中间件设置======================================================
	if gin.Mode() == gin.DebugMode {
		//控制台打印,其实生产可以不适用，跟自定义的日志组件有重复，debug可以测试查看
		//r.Use(gin.Logger())
		//参数打印
		r.Use(ginmiddleware.GinPrintParams())
	}
	//全局跨域中间件设置，某些API单独的中间件禁止全局使用
	r.Use(ginmiddleware.GinCors())

	//全局自定义日志中间件设置,每24小时切割,某些API单独的中间件禁止全局使用,控制台不再打印
	r.Use(ginmiddleware.GinLogger(cfg.LogCfg.LogPath, cfg.LogCfg.LogName, cfg.LogCfg.LogLevel))

	//===================================================中间件设置======================================================

	//==================================================接口路由定义======================================================

	//swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//具体接口实例化
	var apiv1 api.API = v1.NewAPIV1() //v1版本

	//v1版本
	apiV1 := r.Group("/v1")
	{
		//usdt支持
		//签名
		//apiV1.POST("/createusdt", apiv1.CreateUsdtTpl)

		//广播
		apiV1.POST("/push", apiv1.SendTx)

	}
	//==================================================接口路由定义======================================================

	//==================================================定时任务设置======================================================

	if cfg.SystemModel == "cold" {
		//冷系统启动监听文件夹
		//任务定时器查看路由
		//ginmiddleware.AddJobRunner(r)

		//启动测试任务
		//ginmiddleware.JobRunnerDemo()
		//监听文件夹,自动加载新文件地址,
		//var loadService loadkeyservice.BasicService = new(loadkeyservice.LoadService)
		//jobrunner.LoadKeyJob(cfg.CronCfg.LoadKeyJob, cfg.BtcCfg.FilePath, loadService.ReadFile)

		//签名模板
		apiV1.POST("/create", apiv1.CreateTpl)
		//签名
		apiV1.POST("/sign", apiv1.SignTx)

		//私钥验证
		apiV1.GET("/pk", apiv1.GetPrivkey)

		//临时私钥导入
		apiV1.POST("/importpk", apiv1.ImportAddr)

		//创建地址
		apiV1.POST("/createaddr", apiv1.CreateAddrs)

		//

	}
	//==================================================定时任务设置======================================================

}
