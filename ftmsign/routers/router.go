package routers

import (
	"fmt"
	"github.com/eth-sign/conf"
	"github.com/gin-gonic/gin"

	"github.com/eth-sign/routers/apis"
	"github.com/eth-sign/util"
	"log"
	//log "log"
	"net/http"
	"strings"
)

func InitRouters(group *gin.RouterGroup) {

	api := apis.CreateApis()

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

		group.POST("/createaddr", api.CreateAddress)            //done 已验证
		group.POST("/getBalance", api.GetBalance)               //done 已验证
		group.POST("/validAddress", api.ValidAddress)           //done 已验证
		group.POST("/sign", api.Sign)                           //done 已验证
		group.POST("/transfer", api.Transfer)                   //done 已验证
		group.POST("/transferWithNonce", api.TransferWithNonce) //done 已验证
	}

}

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authInfo := c.Request.Header.Get("Authorization")
		if !conf.Config.AuthCfg.Enable {
			log.Println("auth 认证未开启")
			c.Next()
			return
		}
		if len(authInfo) == 0 {
			log.Println("没有传输authinfo信息")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
			return
		}
		authStr := util.DecodeBasicAuth(authInfo)
		userInfo := strings.Split(authStr, ":")
		username := conf.Config.AuthCfg.User
		password := conf.Config.AuthCfg.Password
		if userInfo[0] == username && userInfo[1] == password {
			log.Println("auth 验证通过")
			c.Next()
			return
		}
		log.Println("auth验证不通过")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
		c.Next()
	}
}
