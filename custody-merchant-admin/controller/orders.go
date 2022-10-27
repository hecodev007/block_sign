package controller

import (
	"bytes"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

// UpdatePassOrderInfo
// 订单通过
func UpdatePassOrderInfo(c *handler.Context) error {
	up := new(domain.UpdateOrders)
	err := c.DefaultBinder(up)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	_, err = service.UpdatePassOrderService(up.Id, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessPass)
	return res.ResultOk(c)
}

// UpdateAllOrder
// 订单管理
// 更新订单：一键通过
func UpdateAllOrder(c *handler.Context) error {

	up := new(domain.SelectOrderInfo)
	err := c.DefaultBinder(up)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	total, err := service.UpdateOrderAllService(up, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessPass)
	res.AddData("total", total)
	return res.ResultOk(c)
}

// UpdateThawOrderInfo
// 订单解冻
func UpdateThawOrderInfo(c *handler.Context) error {
	up := new(domain.UpdateOrders)
	err := c.DefaultBinder(up)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	err = service.UpdateThawOrderService(up, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessThaw)
	return res.ResultOk(c)
}

// UpdateFreezeOrderInfo
// 订单冻结
func UpdateFreezeOrderInfo(c *handler.Context) error {
	up := new(domain.UpdateOrders)
	err := c.DefaultBinder(up)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	err = service.UpdateFreezeOrderService(up, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessFreeze)
	return res.ResultOk(c)
}

// UpdateRefuseOrderInfo
// 订单拒绝
func UpdateRefuseOrderInfo(c *handler.Context) error {
	up := new(domain.UpdateOrders)
	err := c.DefaultBinder(up)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if user.MerchantId == 0 {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	err = service.UpdateRefuseOrderService(up, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessRefuse)
	return res.ResultOk(c)
}

// FindOrderList
// 订单管理
// 查询订单列表
func FindOrderList(c *handler.Context) error {
	w := new(domain.SelectOrderInfo)
	w.ServiceId = xkutils.StrToInt(c.QueryParam("service_id"))
	w.Contents = c.QueryParam("contents")
	w.CoinId = xkutils.StrToInt(c.QueryParam("coin_id"))
	w.SerialNo = c.QueryParam("serial_no")
	w.ChainName = c.QueryParam("chain_name")
	w.StartTime = c.QueryParam("start_time")
	w.EndTime = c.QueryParam("end_time")
	w.OrderResult = xkutils.StrToInt(c.QueryParam("order_result"))
	w.Offset, w.Limit = c.OffsetPage()
	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	find, total, err := service.FindOrderListService(w, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", find)
	res.AddData("total", total)
	return res.ResultOk(c)
}

// FindOrderExport
// 订单管理
// 导出订单列表
func FindOrderExport(c *handler.Context) error {
	w := new(domain.SelectOrderInfo)
	w.ServiceId = xkutils.StrToInt(c.QueryParam("service_id"))
	w.Contents = c.QueryParam("contents")
	w.CoinId = xkutils.StrToInt(c.QueryParam("coin_id"))
	w.SerialNo = c.QueryParam("serial_no")
	w.ChainName = c.QueryParam("chain_name")
	w.StartTime = c.QueryParam("start_time")
	w.EndTime = c.QueryParam("end_time")
	w.OrderResult = xkutils.StrToInt(c.QueryParam("order_result"))
	w.Offset, w.Limit = c.OffsetPage()
	user := c.GetTokenUser()
	if user.MerchantId == 0 && !user.Admin {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	find, err := service.FindOrderExportService(w, user.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	name := fmt.Sprintf("订单-%d.xlsx", time.Now().Local().Unix())
	//设置请求头 使用浏览器下载
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+name)
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, bytes.NewReader(find.Bytes()))
}

// CountOrderStatus
// 订单管理
// 统计各个订单数量
func CountOrderStatus(c *handler.Context) error {
	user := c.GetTokenUser()
	id := user.MerchantId
	if user.MerchantId == 0 {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	find, err := service.CountOrderStatusService(id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", find)
	return res.ResultOk(c)
}

// FinOrderPlanDetail
// 订单管理
// 查看订单详情
func FinOrderPlanDetail(c *handler.Context) error {

	id := xkutils.StrToInt64(c.QueryParam("id"))
	user := c.GetTokenUser()
	if user.MerchantId == 0 {
		return handler.NewError(c, global.OperationIsNotAuditErr)
	}
	detail, err := service.FindOrderDetailService(id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", detail)
	return res.ResultOk(c)
}

func OrderRollBack(c *handler.Context) error {
	req := new(domain.OrderOperateReq)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	_, err = service.OrderRollBack(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}
