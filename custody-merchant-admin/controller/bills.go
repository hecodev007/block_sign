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

// CreateBillDetail
// 创建账单
func CreateBillDetail(c *handler.Context) error {

	w := new(domain.BillInfo)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	w.CreateByUser = user.Id
	//err = service.CreateWithdrawBill(w)
	//if err != nil {
	//	return financeHandler.NewError(c, err.Error())
	//}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// UpdateBillDetail
// @Tags 账单管理
// @Summary  更新账单
// @Description 更新账单
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param body body domain.BillInfo true "参数"
// @Param select_time query string false "查询时间"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/bill/updateBillDetail [post]
func UpdateBillDetail(c *handler.Context) error {
	w := new(domain.BillInfo)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	w.CreateByUser = user.Id
	total, err := service.UpdateBillDetailService(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	return res.ResultOk(c)
}

// PushBill
// @Tags 账单管理
// @Summary  重推账单
// @Description 重推账单
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param body body domain.BillInfo true "参数"
// @Param select_time query string false "查询时间"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/bill/pushBill [post]
func PushBill(c *handler.Context) error {

	w := new(domain.BillInfo)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = service.PushBillService(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessPush)
	return res.ResultOk(c)
}

// FindBillList
// @Tags 账单管理
// @Summary  查询账单
// @Description 查询账单
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param coin_id query int64 false "币种Id参数"
// @Param service_id query  int false  "业务线Id参数"
// @Param address query  string false "地址参数"
// @Param tx_type query  int false  "交易类型Id参数"
// @Param min query  float64 false "最大值参数"
// @Param max query  float64 false  "最小值参数"
// @Param create_time query string false "时间参数"
// @Param limit query  int  false "查询条数"
// @Param offset query int  false "查询起始位置"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/bill/list [get]
func FindBillList(c *handler.Context) error {

	w := new(domain.BillSelect)
	w.MerchantId = xkutils.StrToInt64(c.QueryParam("merchant_id"))
	w.Phone = c.QueryParam("phone")
	w.ServiceId = xkutils.StrToInt(c.QueryParam("service_id"))
	w.Address = c.QueryParam("address")
	w.BillStatus = xkutils.StrToInt(c.QueryParam("bill_status"))
	w.Offset, w.Limit = c.OffsetPage()
	w.TxStartTime = c.QueryParam("tx_start_time")
	w.TxEndTime = c.QueryParam("tx_end_time")
	w.ConfirmStartTime = c.QueryParam("confirm_start_time")
	w.ConfirmEndTime = c.QueryParam("confirm_end_time")
	bill, total, err := service.FindBillService(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", bill)
	return res.ResultOk(c)
}

// BillExcelExport
// @Tags 账单管理
// @Summary  导出账单
// @Description 导出账单
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param coin_id query int64 false "币种Id参数"
// @Param service_id query  int false  "业务线Id参数"
// @Param address query  string false "地址参数"
// @Param tx_type query  int false  "交易类型Id参数"
// @Param min query  float64 false "最大值参数"
// @Param max query  float64 false  "最小值参数"
// @Param create_time query string false "时间参数"
// @Param limit query  int  false "查询条数"
// @Param offset query int  false "查询起始位置"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/bill/billExcelExport [get]
func BillExcelExport(c *handler.Context) error {
	w := new(domain.BillSelect)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	buff, err := service.ExportBillService(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	name := fmt.Sprintf("%s.xlsx", time.Now().Local().Format("2006-01-02 15:04:05.9999"))
	//设置请求头 使用浏览器下载
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+name)
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, bytes.NewReader(buff.Bytes()))
}

// FindBillBalance
// @Tags 账单管理
// @Summary  查询账单的金额
// @Description 查询账单的金额
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param unit_id query int false "单位Id参数"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/bill/findBillBalance [get]
func FindBillBalance(c *handler.Context) error {
	w := new(domain.BillBalance)
	user := c.GetTokenUser()
	uid := user.Id
	if user.Admin {
		uid = 0
	}
	w.UnitId = xkutils.StrToInt(c.QueryParam("unit_id"))
	buff, err := service.FindBillBalanceService(w, uid)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", buff)
	return res.ResultOk(c)
}
