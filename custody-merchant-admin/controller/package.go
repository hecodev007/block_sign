package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/router/web/handler"
)

// CreateNewPackage 增加套餐
func CreateNewPackage(c *handler.Context) error {

	req := new(domain.PackageInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.CreatePackageItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)

}

// DeletePackageItem 删除套餐
func DeletePackageItem(c *handler.Context) error {

	req := new(domain.PackageReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.DeletePackageItem(user.Id, req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// UpdatePackageItem 修改套餐
func UpdatePackageItem(c *handler.Context) error {
	var req = map[string]interface{}{}
	err := c.Binder(&req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	_, ok := req["id"]
	if !ok {
		return handler.NewError(c, global.DataWarnNoDataErr)
	}

	user := c.GetTokenUser()
	err = service.UpdatePackageItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// SearchPackageList 查询套餐列表
func SearchPackageList(c *handler.Context) error {

	req := new(domain.PackageReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	req.Offset, req.Limit = OffsetPage(req.Limit, req.Offset)
	//user := c.GetTokenUser()
	list, total, err := service.SearchPackages(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)
}

// SearchPackageItemInfo 套餐详情
func SearchPackageItemInfo(c *handler.Context) error {

	req := new(domain.PackageReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	item, err := service.SearchPackageItem(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(item)
	return res.ResultOk(c)
}

// SearchMchPackageItemInfo 商户查询套餐详情
func SearchMchPackageItemInfo(c *handler.Context) error {

	req := new(domain.MchPackageReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	log.Errorf("SearchMchPackageItemInfo req= %+v", req)
	item, err := service.SearchMchPackageItem(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(item)
	log.Errorf("SearchMchPackageItemInfo res= %+v", res)
	return res.ResultOk(c)
}

// SearchPackageScreenList 筛选列表
func SearchPackageScreenList(c *handler.Context) error {

	req := new(domain.PackageReqInfo)
	req.Screen = c.QueryParam("screen")

	//user := c.GetTokenUser()
	typeList, tradeList, modelList, err := service.SearchPackageScreen(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("type_list", typeList)
	res.AddData("trade_list", tradeList)
	res.AddData("model_list", modelList)
	return res.ResultOk(c)
}
