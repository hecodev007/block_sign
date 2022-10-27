package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/celo-sign/conf"
	"github.com/group-coldwallet/celo-sign/routers/apis"
	"github.com/group-coldwallet/celo-sign/util"
	log "github.com/sirupsen/logrus"
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

		group.POST("/createaddr", api.CreateAddress)
		group.POST("/sign", api.Sign)
		if conf.Config.WalletType == "hot" {
			group.POST("/transfer", api.Transfer)
		}
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
