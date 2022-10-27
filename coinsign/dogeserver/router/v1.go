package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/dogeserver/api/v1"
	"github.com/group-coldwallet/dogeserver/conf"
)

func InitV1Router(r *gin.Engine) {
	v1Router := r.Group("/v1")
	{
		v1Router.POST("/transfer", gin.BasicAuth(gin.Accounts{
			"rylink2020": "rylinkhoo2020", //用户名：密码
		}), v1.Transfer)
		if conf.GlobalConf.SystemModel == "cold" {
			v1Router.POST("/createaddr", v1.CreateAddrs)
		}
	}
}
