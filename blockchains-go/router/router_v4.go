package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/middleware"
	v3 "github.com/group-coldwallet/blockchains-go/router/api/v3"
)

//验证签名的接口全部使用Content-Type: application/x-www-form-urlencoded
func InitV4Router(r *gin.Engine) {
	v3Router := r.Group("/v4")
	{
		// 新版出账（和上面相比只是验签不一样）
		v3Router.POST("/applyTransaction", middleware.CheckApiSignSecure(), middleware.AuthApiV4(), v3.Transfer)
	}
}
