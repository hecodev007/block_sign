package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
	"fmt"
)

// GetMainUserList
// 人员管理
// 获取人员列表
func GetMainUserList(c *handler.Context) error {
	u := new(domain.SelectUserInfo)
	u.Account = c.QueryParam("account")
	u.MerchantId = c.SwitchType("merchant_id", "int64").(int64)
	u.Sid = c.SwitchType("sid", "int").(int)
	u.Offset, u.Limit = c.OffsetPage()
	res := handler.NewSuccess()
	if u.MerchantId == 0 && u.Account == "" {
		res.AddData("merchant", map[string]interface{}{})
		res.AddData("list", []domain.SelectUserList{})
		res.AddData("total", 0)
		return res.ResultOk(c)
	}

	// 先查询商家信息
	mainInfo, err := service.GetMainUserInfoService(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	// 商户信息是否存在
	if mainInfo.Id <= 0 {
		return handler.NewError(c, global.DataWarnNoDataErr)
	}
	u.MerchantId = mainInfo.Id
	list, total, err := service.FindServiceChainsByMid(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	res.AddData("merchant", mainInfo)
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

func GetMainUserById(c *handler.Context) error {

	merchantId := c.SwitchType("merchant_id", "int64").(int64)
	subInfo, err := service.GetUserInfoById(merchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", subInfo)
	return res.ResultOk(c)
}

func GetChainsInfo(c *handler.Context) error {
	sId := c.SwitchType("service_id", "int").(int)
	info, err := service.GetServiceChainsInfo(sId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", info)
	return res.ResultOk(c)
}

func GetRoleInfo(c *handler.Context) error {
	id := c.SwitchType("service_id", "int").(int)
	fmt.Println(id)
	info, err := service.GetServiceChainsRolesInfo(id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", info)
	return res.ResultOk(c)
}
