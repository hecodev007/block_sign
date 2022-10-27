package middleware

import (
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/cocos/common"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authInfo := c.Request.Header.Get("Authorization")
		if !viper.GetBool("auth.enable") {
			logs.Debug("auth 认证未开启")
			c.Next()
			return
		}
		if len(authInfo) == 0 {
			logs.Debug("没有传输authinfo信息")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
			return
		}
		authStr := common.DecodeBasicAuth(authInfo)
		userInfo := strings.Split(authStr, ":")
		username := viper.GetString("auth.username")
		password := viper.GetString("auth.password")
		if userInfo[0] == username && userInfo[1] == password {
			logs.Debug("auth 验证通过")
			c.Next()
			return
		}
		logs.Debug("auth验证不通过")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "401", "message": "Unauthorized"})
		c.Next()
	}
}
