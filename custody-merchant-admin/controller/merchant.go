package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
)

// SearchApplyList 商户申请列表
func SearchApplyList(c *handler.Context) error {

	req := new(domain.ApplyReqInfo)
	err := c.Binder(req)
	req.ContactStr = c.QueryParam("contact_str")
	req.AccountId = c.QueryParam("account_id")
	req.AccountName = c.QueryParam("account_name")
	req.CardNum = c.QueryParam("card_num")
	req.VerifyStatus = c.QueryParam("verify_status")
	req.VerifyResult = c.QueryParam("verify_result")
	req.Offset, req.Limit = c.OffsetPage()

	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	list, total, err := service.SearchApplyList(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

//ActionMerchantItem 操作商户申请，通过/拒绝
func ActionMerchantItem(c *handler.Context) error {

	req := new(domain.MerchantOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.OperateApplyItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//SearchMerchantList 商户列表(已经通过申请的商户)
func SearchMerchantList(c *handler.Context) error {

	req := new(domain.MerchantReqInfo)
	err := c.DefaultBinder(req)
	req.ContactStr = c.QueryParam("contact_str")
	req.AccountId = c.QueryParam("account_id")
	req.AccountName = c.QueryParam("account_name")
	req.CardNum = c.QueryParam("card_num")
	req.RealNameStatus = c.QueryParam("real_name_status")
	req.FvStatus = c.QueryParam("fv_status")
	req.LockStatus = c.QueryParam("lock_status")
	req.RealNameStart = c.QueryParam("real_name_start")
	req.RealNameEnd = c.QueryParam("real_name_end")
	req.Offset, req.Limit = c.OffsetPage()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	list, total, err := service.SearchMerchantList(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

func GetApplyImageInfo(c *handler.Context) error {
	req := new(domain.MerchantImageInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	image, err := service.SearchApplyImageInfo(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(image)
	return res.ResultOk(c)
}

//GetMerchantImageInfo 获取认证图片/合同详情
func GetMerchantImageInfo(c *handler.Context) error {
	req := new(domain.MerchantImageInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	image, err := service.SearchMerchantImageInfo(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(image)
	return res.ResultOk(c)
}

//SearchMerchantInfo 商户编辑详情
func SearchMerchantInfo(c *handler.Context) error {

	req := new(domain.MerchantOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	image, err := service.SearchMerchantInfo(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccessByStruct(image)
	return res.ResultOk(c)
}

//UpdateMerchantItem 编辑商户
func UpdateMerchantItem(c *handler.Context) error {

	req := map[string]interface{}{}
	err := c.Binder(&req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	_, ok := req["id"]
	if !ok {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	user := c.GetTokenUser()
	err = service.UpdateMerchantItem(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//PushMerchantItem 推送财务审核
func PushMerchantItem(c *handler.Context) error {

	req := new(domain.MerchantOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	err = service.PushMerchantToFinanceVerify(req.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//PushMerchantAll 一键推送财务审核
func PushMerchantAll(c *handler.Context) error {

	req := new(domain.BusinessInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	err = service.PushBatchApplysToFinanceVerify()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//UpdateMerchantImage 商户编辑详情
func UpdateMerchantImage(c *handler.Context) error {

	req := new(domain.MerchantImgInfo)
	err := c.DefaultBinder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	// user := c.GetTokenUser()
	err = service.UpdateMerchantImage(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, "上传成功")
	return res.ResultOk(c)
}
