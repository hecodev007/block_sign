package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
)

//SearchFinanceList 财务审核列表
func SearchFinanceList(c *handler.Context) error {

	req := new(domain.MerchantReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	list, total, err := service.SearchPushApplyList(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)
}

//ActionFinanceItem 操作财务审核申请，解冻冻结资产/解冻冻结
func ActionFinanceItem(c *handler.Context) error {

	req := new(domain.FinanceOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.UpdateFinanceLock(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func FinanceAgreeRefuse(c *handler.Context) error {

	req := new(domain.FinanceOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.FinanceAgreeRefuse(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func FinanceItemImg(c *handler.Context) error {

	req := new(domain.ApplyImageReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	data, err := service.GetFinanceVerifyImage(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(data)
	return res.ResultOk(c)
}

//UpdateFinanceItem 编辑商户（认证图片/合同图片/时间）
func UpdateFinanceItem(c *handler.Context) error {

	var req map[string]interface{}
	err := c.Binder(&req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	_, ok := req["id"]
	if !ok {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}

	user := c.GetTokenUser()
	err = service.UpdateFinanceVerifyImage(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//FinanceLockLogList 冻结日志详情
func FinanceLockLogList(c *handler.Context) error {

	req := new(domain.MerchantReqInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	list, total, err := service.SearchFinanceLockRecordList(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", list)
	return res.ResultOk(c)
}
