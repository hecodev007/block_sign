package main

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/cocos/api/cocos"
	"github.com/group-coldwallet/cocos/launcher"
	"github.com/group-coldwallet/cocos/middleware"
	"github.com/spf13/viper"
)

func main() {
	/*获取配置文件*/
	launcher.InitConfig()
	launcher.CustomLogInfo()
	launcher.UnlockWallet()
	launcher.InitDB()
	r := gin.Default()
	/*中间件 auth认证*/
	r.Use(middleware.BasicAuth())

	cocos := cocos.Cocos{}
	v1 := r.Group("/v1")
	cocosbcx := v1.Group("/cocosbcx")
	{
		//注册账户  经过沟通，暂时不再使用
		cocosbcx.POST("/createaccount", cocos.CreateAccount)
		//账户余额
		cocosbcx.GET("/getblance", cocos.GetBalance)
		//转账
		cocosbcx.POST("/transfer", cocos.Transfer)
		//验证地址
		cocosbcx.GET("/vaildaddress", cocos.VaildAddress)
	}
	r.Run(viper.GetString("http.listen_address"))
}
