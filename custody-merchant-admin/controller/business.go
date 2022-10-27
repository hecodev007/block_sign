package controller

import (
	"bytes"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
	"encoding/json"
	"io/ioutil"
)

// CreateNewBusiness 增加业务线
func CreateNewBusiness(c *handler.Context) error {
	//获取完之后要重新设置进去，不然会丢失
	data, _ := ioutil.ReadAll(c.Request().Body)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(data))
	reqS := new(domain.CreateBusinessInfo)
	err := c.Binder(reqS)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	var reqM map[string]interface{}
	json.Unmarshal(data, &reqM)
	user := c.GetTokenUser()
	err = service.CreateBusinessItem(user, reqS, &reqM)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)

}

// DeleteBusinessItem 删除业务线
func DeleteBusinessItem(c *handler.Context) error {

	req := new(domain.BusinessInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.DeleteBusinessItem(user, req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)

}

// UpdateBusinessItem 修改业务线
func UpdateBusinessItem(c *handler.Context) error {

	var req = map[string]interface{}{}
	err := c.Binder(&req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	_, ok := req["id"]
	if !ok {
		return handler.NewError(c, global.DataWarnNoDataErr)
	}
	_, ok = req["trade_type"]
	if !ok {
		return handler.NewError(c, global.DataWarnNoTransferTypeErr)
	}
	user := c.GetTokenUser()
	err = service.UpdateBusinessItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// SearchBusinessList 查询业务线列表
func SearchBusinessList(c *handler.Context) error {

	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	req.Offset, req.Limit = OffsetPage(req.Limit, req.Offset)

	//user := c.GetTokenUser()
	list, total, err := service.SearchBusiness(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)

}

// SearchBusinessItemInfo 查询业务线详情
func SearchBusinessItemInfo(c *handler.Context) error {

	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	item, err := service.SearchBusinessItem(req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(item)
	return res.ResultOk(c)

}

// ActionBusinessItem 操作业务线，冻结/解冻
func ActionBusinessItem(c *handler.Context) error {

	req := new(domain.BusinessOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.OperateBusinessItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// BusinessPackageInfo 套餐费用详情接口
func BusinessPackageInfo(c *handler.Context) error {

	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	//item, err := service.SearchBusinessPackageInfo(req.Id)
	//if err != nil {
	//	return financeHandler.NewError(c, err.Error())
	//}
	//res := financeHandler.NewSuccessByStruct(item)
	//return res.ResultOk(c)
	return err
}

// BusinessSecurity 安全信息
func BusinessSecurity(c *handler.Context) error {
	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	item, err := service.SearchBusinessSecurityInfo(req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(item)
	return res.ResultOk(c)

}

// BusinessOperateLogList 操作日志列表
func BusinessOperateLogList(c *handler.Context) error {

	req := new(domain.RecordReqInfo)
	//req.Offset, req.Limit = c.OffsetPage()
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	list, total, err := service.SearchRecords(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)

}

// GetClientIdAndSecret 获取密钥CLIENT_ID/SECRET
func GetClientIdAndSecret(c *handler.Context) error {

	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	var cid string
	var secret string
	cid, secret, err = service.GetNewAccoutInfo(*req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("client_id", cid)
	res.AddData("secret", secret)
	return res.ResultOk(c)

}

// ResetClientIdAndSecret 重置密钥CLIENT_ID/SECRET
func ResetClientIdAndSecret(c *handler.Context) error {

	req := new(domain.BusinessReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	var bsInfo domain.BusinessSecurityInfo
	bsInfo, err = service.ResetBusinessClientIdAndSecret(req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("client_id", bsInfo.ClientId)
	res.AddData("secret", bsInfo.Secret)
	return res.ResultOk(c)

}
