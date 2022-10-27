package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// IP 访问白名单
func IPWhiteListMiddleware(ipList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//上传代码关闭，注意填写
		flag := false
		clientIp := c.ClientIP()
		for _, tmp := range ipList {
			host := tmp
			if clientIp == host {
				flag = true
				break
			}
		}
		if !flag {
			//c.String(401, "%s, 不在白名单中拒绝访问 \n", clientIp)
			c.String(401, fmt.Sprintf("refuse %s", clientIp))
			c.Abort()
		}

	}
}

// IP 访问黑名单
func IPBlackListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipList := []string{
			"127.0.0.1",
		}
		flag := false
		clientIp := c.ClientIP()
		for _, host := range ipList {
			if clientIp == host {
				flag = true
				break
			}
		}
		if flag {
			//c.String(401, "%s, 在黑名单中，拒绝访问 \n", clientIp)
			c.String(401, "拒绝访问 \n", clientIp)
			c.Abort()
		}
	}
}
