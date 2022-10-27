package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/middleware"
	"github.com/group-coldwallet/blockchains-go/router/api/custody"
)

//验证签名的接口全部使用Content-Type: application/x-www-form-urlencoded
func InitCustodyRouter(r *gin.Engine) {
	custodyRouter := r.Group("/custody")
	{
		//middleware.CheckApiParamSign()
		//主链币/代币 列表
		//custodyRouter.POST("/coin/list",custody.GetCoinList)
		////商户 创建client_id,安全secret
		//custodyRouter.POST("/create/mch", custody.CreateClientIdSecret)
		////商户重置密钥 client_id,安全secret
		//custodyRouter.POST("/resecret/mch", custody.ReSecretClientIdSecret)
		////商户查询 client_id,安全secret
		//custodyRouter.POST("/get/mch",custody.SearchClientIdSecret)
		////验证托管后台接受来自商户的参数
		//custodyRouter.POST("/verify/param", custody.VerifyParamFromCustody)
		////创建地址回传
		//custodyRouter.POST("/address/back", custody.BackAddress)
		////商户钱包地址创建（批量创建地址）
		//custodyRouter.POST("/address",custody.BatchCreateAddress)
		////商户钱包地址创建（多币种创建地址）
		//custodyRouter.POST("/lot/coin/address", custody.BatchCreateLotCoinAddress)
		////商户地址绑定（充值回调地址） 批量
		//custodyRouter.POST("/address/bind", custody.BindAddress)
		////提现接口
		//custodyRouter.POST("/withdraw",custody.OperateBalance)
		//
		//middleware.CheckApiParamSign()
		//主链币/代币 列表
		custodyRouter.POST("/coin/list", middleware.CheckApiParamSign(), custody.GetCoinList)
		//商户 创建client_id,安全secret
		custodyRouter.POST("/create/mch", middleware.CheckApiParamSign(), custody.CreateClientIdSecret)
		//商户重置密钥 client_id,安全secret
		custodyRouter.POST("/resecret/mch", middleware.CheckApiParamSign(), custody.ReSecretClientIdSecret)
		//商户查询 client_id,安全secret
		custodyRouter.POST("/get/mch", middleware.CheckApiParamSign(), custody.SearchClientIdSecret)
		//验证托管后台接受来自商户的参数
		custodyRouter.POST("/verify/param", middleware.CheckApiParamSign(), custody.VerifyParamFromCustody)
		//创建地址回传
		custodyRouter.POST("/address/back", custody.BackAddress)
		//商户钱包地址创建（批量创建地址）
		custodyRouter.POST("/address", middleware.CheckApiParamSign(), custody.BatchCreateAddress)
		//商户钱包地址创建（多币种创建地址）
		custodyRouter.POST("/lot/coin/address", middleware.CheckApiParamSign(), custody.BatchCreateLotCoinAddress)
		//商户地址绑定（充值回调地址） 批量
		custodyRouter.POST("/address/bind", middleware.CheckApiParamSign(), custody.BindAddress)
		//提现接口
		custodyRouter.POST("/withdraw", middleware.CheckApiParamSign(), custody.OperateBalance)

		////余额查询 TODO:未完成
		custodyRouter.POST("/coin/balance",custody.CoinBalance)
		//上链结果回调/查询
		custodyRouter.POST("/upchain/status", middleware.CheckApiParamSign(), custody.OrderUpChainStatus)

	}
}
