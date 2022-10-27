package web

import (
	"custody-merchant-admin/financeController"
	"github.com/labstack/echo/v4"
)

func financeRouter(e *echo.Echo) {
	f := e.Group("/finance/open/v1")
	{
		// 获取商户中属于财务的资产
		f.GET("/merchants/assets/list", handler(financeController.FindFinanceAssetsList))

		// 通过/拒绝商户
		f.POST("/operation/merchants", handler(financeController.PassOrRefuseMerchant))
		// 冻结资产
		f.POST("/freeze/assets", handler(financeController.FreezeMerchantAssets))
		// 冻结账户
		f.POST("/freeze/account", handler(financeController.FreezeMerchantAccount))
		// 资料推送
		f.POST("/push/merchants", handler(financeController.PushMerchant))
	}
}
