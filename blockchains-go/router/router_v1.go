package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/middleware"
	"github.com/group-coldwallet/blockchains-go/router/api/v1"
)

func InitV1Router(r *gin.Engine) {
	//
	v1Router := r.Group("/v1")
	{
		//v1Router.POST("/signature", v1.Signature)
		//v1Router.POST("/applyAddress", v1.ApplyAddress)
		v1Router.POST("/auto/repush", v1.RePushTx)
		v1Router.POST("/applyAddress", v1.ApplyAddress)
		v1Router.GET("/GetCoinSet", v1.GetCoinList)
		v1Router.GET("/MchCoinSum", v1.GetMchBalance)

		//middleware.CheckSian() 校验签名
		//middleware.AuthApi() api 权限认证，该路径需要传入币种名
		v1Router.POST("/applyTransaction", middleware.CheckSign2(), middleware.AuthApi(), v1.Transfer)
		//v1Router.POST("/applyTransaction", middleware.AuthApi(), v1.Transfer)
		v1Router.GET("/transPushTest", v1.TransPushTest)
		v1Router.GET("/get_transaction_list", v1.FindTransactionList)

		//目前v1只使用该接口
		v1Router.POST("/transPush", gin.BasicAuth(gin.Accounts{
			"rylink": "fdhj&%@#13*74", //用户名：密码
		}), v1.TransPush)

	}

}
