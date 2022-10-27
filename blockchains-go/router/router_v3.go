package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/middleware"
	v3 "github.com/group-coldwallet/blockchains-go/router/api/v3"
)

//验证签名的接口全部使用Content-Type: application/x-www-form-urlencoded
func InitV3Router(r *gin.Engine) {
	v3Router := r.Group("/v3")
	{
		//walletserver程序回调
		v3Router.POST("/walletserver/callback", v3.WalletCallBack)
		v3Router.POST("/walletserver/multifrom/callback", v3.CallbackMulAddrTrx)

		//交易所调用接口
		v3Router.POST("/MchCoinSum", middleware.CheckApiSign(), middleware.AuthApiV3(), v3.GetMchBalance)

		// 查询所有币种的余额
		v3Router.POST("/MchCoinListSum", v3.GetMchAllBalanceV2)

		//v3Router.POST("/trxFix", v3.TrxFix)
		//v3Router.POST("/trxRePush", v3.TrxRePush)

		// 获取单个币种出账地址最大余额
		v3Router.POST("/MchCoinMaxBalance", middleware.CheckApiSign(), v3.GetMchCoinMaxBalance)

		// 生成地址
		v3Router.POST("/applyAddress", middleware.CheckApiSign(), middleware.AuthApiV3(), v3.ApplyAddress)

		// 出账
		v3Router.POST("/applyTransaction", middleware.CheckApiSign(), middleware.AuthApiV3(), v3.Transfer)

		// 新版出账（和上面相比只是验签不一样）
		v3Router.POST("/applyTransactionSecure", v3.Transfer)
		v3Router.GET("/ethgas", v3.GetEthEstFee)

		// 查询订单状态
		v3Router.POST("/findOrderStatus", v3.FindOrderStatus)

		// 校验地址是否合法
		v3Router.POST("/validAddress", middleware.CheckApiSign(), v3.ValidAddress)
		v3Router.POST("/isInsideAddress", middleware.CheckApiSign(), v3.ValidInsideAddress)
	}
}
