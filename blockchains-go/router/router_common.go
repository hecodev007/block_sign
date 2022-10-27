package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/router/api"
	v3 "github.com/group-coldwallet/blockchains-go/router/api/v3"
	"github.com/group-coldwallet/blockchains-go/runtime/job"
)

//基本不会有变动的路由
func InitCommonRouter(r *gin.Engine, tm *job.TxManager) {

	//==================定时任务路由==================
	//定时任务json结构显示
	r.GET("/jobrunner/json", api.JobJson)
	//定时任务静态界面文件指定
	r.LoadHTMLGlob("resource/views/status.html")
	//根据上方加载的模板返回给定端点处的html页面
	r.GET("/jobrunner/html", api.JobHtml)
	//==================定时任务路由==================

	//==================钉钉通知路由==================
	//r.POST("/blockchains/ding", v1.DingOutgoing)
	r.POST("/blockchains/ding", func(ctx *gin.Context) {
		v3.DingOutgoing(ctx, tm)
	})
	//==================钉钉通知路由==================

}
