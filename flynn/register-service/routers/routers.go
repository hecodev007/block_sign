package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/flynn/register-service/apis"
	"github.com/group-coldwallet/flynn/register-service/conf"
	"github.com/group-coldwallet/flynn/register-service/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

/*
定义路由
*/
func InitRouters(group *gin.RouterGroup) {

	group.Use(BasicAuth())
	{
		group.GET("ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"message": "success",
				"data":    "online",
			})
		})
		//插入地址接口
		group.POST("address/insert", apis.InsertWatchAddress)
		//删除地址
		//插入合约接口
		group.POST("contract/insert", apis.InsertContractInfo)
		//删除合约
		//更新地址
		//推补数据
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
