package controller

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func FindChainBillList(c *handler.Context) error {

	w := new(domain.ChainBillSelect)
	w.MerchantId = xkutils.StrToInt64(c.QueryParam("merchant_id"))
	w.Phone = c.QueryParam("phone")
	w.AddressOrMemo = c.QueryParam("address_or_memo")
	w.TxType = xkutils.StrToInt(c.QueryParam("tx_type"))
	w.IsReback = xkutils.StrToInt(c.QueryParam("is_reback"))
	w.StartTime = c.QueryParam("start_time")
	w.EndTime = c.QueryParam("end_time")
	w.ConfirmStartTime = c.QueryParam("confirm_start_time")
	w.ConfirmEndTime = c.QueryParam("confirm_end_time")
	w.Offset, w.Limit = c.OffsetPage()
	bill, total, err := service.FindChainBillService(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("total", total)
	res.AddData("list", bill)
	return res.ResultOk(c)
}

func FindChainBillExport(c *handler.Context) error {

	w := new(domain.ChainBillSelect)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	w.Offset, w.Limit = c.OffsetPage()
	bill, err := service.FindChainBillExport(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	name := fmt.Sprintf("链上订单-%d.xlsx", time.Now().Local().Unix())
	//设置请求头 使用浏览器下载
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+name)
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, bytes.NewReader(bill.Bytes()))

}
func FindChainBillReBack(c *handler.Context) error {
	w := new(domain.UpChainInfo)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if w.Id == 0 {
		return handler.NewError(c, "id=0")
	}
	err = service.RollBackChainBill(w.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, "已发送回滚")
	return res.ResultOk(c)

}
func FindChainBillRePush(c *handler.Context) error {
	w := new(domain.UpChainInfo)
	err := c.DefaultBinder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = service.RePushChainBill(w.Id)
	if err != nil {
		return err
	}
	res := handler.NewResult(200, "重推成功")
	return res.ResultOk(c)
}
