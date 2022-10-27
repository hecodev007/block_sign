package controller

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
)

// GetAssetsList
// 资产管理
// 资产管理列表、数据图
func GetAssetsList(c *handler.Context) error {
	as := new(domain.AssetsSelect)
	//t := new(domain.AssetsByTag)
	as.Offset, as.Limit = c.OffsetPage()
	//as.Tag = c.QueryParam("tag")
	//as.StartTime = c.QueryParam("start_time")
	//as.EndTime = c.QueryParam("end_time")
	as.MerchantId = xkutils.StrToInt64(c.QueryParam("merchant_id"))
	as.ServiceId = xkutils.StrToInt(c.QueryParam("service_id"))
	as.ServiceState = xkutils.StrToInt(c.QueryParam("service_state"))
	as.CoinState = xkutils.StrToInt(c.QueryParam("coin_state"))
	as.UnitId = xkutils.StrToInt(c.QueryParam("unit_id"))
	as.IsTest = xkutils.StrToInt(c.QueryParam("is_test"))
	as.Show = xkutils.StrToInt(c.QueryParam("show"))
	// 查询列表
	list, total, err := service.FindAssetsListService(as, as.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//t.Tag = as.Tag
	//t.StartTime = as.StartTime
	//t.EndTime = as.EndTime
	//t.UserId = as.MerchantId
	//// 查询饼图
	//ring, err := service.FindAssetsRingService(as.MerchantId)
	//// 查询折线图
	//line, err := service.FindAssetsByTagService(t)
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)
}

// GetAssetsLine
// 资产管理
// 资产管理列表、数据图
func GetAssetsLine(c *handler.Context) error {
	as := new(domain.AssetsSelect)
	t := new(domain.AssetsByTag)
	//as.Offset, as.Limit = c.OffsetPage()
	as.Tag = c.QueryParam("tag")
	as.StartTime = c.QueryParam("start_time")
	as.EndTime = c.QueryParam("end_time")
	t.Tag = as.Tag
	t.StartTime = as.StartTime
	t.EndTime = as.EndTime
	t.UserId = 0
	// 查询折线图
	line, err := service.FindAssetsByTagService(t)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("line", line)
	return res.ResultOk(c)
}

// GetAssetsRing
// 资产管理
// 资产管理列表、数据图
func GetAssetsRing(c *handler.Context) error {
	// 查询饼图
	ring, err := service.FindAssetsRingService(0)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("ring", ring)
	return res.ResultOk(c)
}
