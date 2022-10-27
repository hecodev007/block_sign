package routers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/bsc-sign/conf"

	"github.com/bsc-sign/routers/apis"
	"github.com/bsc-sign/util"
	"github.com/bsc-sign/util/log"
	"net/http"
	"strings"
)

func InitRouters(ctx context.Context, group *gin.RouterGroup) {

	api := apis.CreateApis(ctx)

	group.Use(BasicAuth())
	{
		group.GET("ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"message": "success",
				"data":    "online",
			})
		})
		group.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"message": "success",
				"data":    fmt.Sprintf("start %s sign service", conf.Config.CoinType),
			})
		})

		group.POST("/createAddr", api.CreateAddress)
		group.POST("/getBalance", api.GetBalance)
		group.POST("/validAddress", api.ValidAddress)
		group.POST("/delkey", api.DelKey)
		group.POST("/sign", api.Sign)
		group.POST("/transfer", api.Transfer)
		group.POST("/transferCollect", api.TransferCollect)
	}

}

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authInfo := c.Request.Header.Get("Authorization")
		if !conf.Config.AuthCfg.Enable {
			log.Info("auth 认证未开启")
			c.Next()
			return
		}
		if len(authInfo) == 0 {
			log.Info("没有传输authinfo信息")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
			return
		}
		authStr := util.DecodeBasicAuth(authInfo)
		userInfo := strings.Split(authStr, ":")
		username := conf.Config.AuthCfg.User
		password := conf.Config.AuthCfg.Password
		if userInfo[0] == username && userInfo[1] == password {
			log.Info("auth 验证通过")
			c.Next()
			return
		}
		log.Info("auth验证不通过")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
		c.Next()
	}
}
