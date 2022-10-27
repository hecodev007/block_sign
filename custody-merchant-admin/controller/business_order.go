package controller

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"

	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/tealeg/xlsx"
	"net/http"
	"strconv"
	"time"
)

//AccountBusinessRenew 业务线续费（商户发起）
func AccountBusinessRenew(c *handler.Context) error {

	req := new(domain.AccountOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	log.Errorf("SearchBusinessOrderList %+v\n", req)
	err = service.RenewBusinessOrder(req.AccountId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	log.Errorf("SearchBusinessOrderList res %+v\n", res)
	return res.ResultOk(c)
}

//SearchBusinessOrderList 订单列表
func SearchBusinessOrderList(c *handler.Context) error {

	req := new(domain.OrderReqInfo)
	req.BusinessId = xkutils.StrToInt(c.QueryParam("business_id"))
	req.AccountId = xkutils.StrToInt64(c.QueryParam("account_id"))
	req.ContactStr = c.QueryParam("contact_str")
	req.OrderId = c.QueryParam("order_id")
	req.Offset, req.Limit = c.OffsetPage()
	//user := c.GetTokenUser()
	log.Infof("SearchBusinessOrderList ")

	log.Infof("SearchBusinessOrderList req:%+v \n", req)
	log.Warnf("SearchBusinessOrderList req:%+v \n", req)
	log.Errorf("SearchBusinessOrderList req:%v \n", req)

	list, total, err := service.SearchBusinessOrderList(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

func DownBusinessOrderList(c *handler.Context) error {

	w := new(domain.OrderReqInfo)
	err := c.Binder(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	w.Offset, w.Limit = OffsetPage(w.Limit, w.Offset)
	bill, err := FindBusinessOrderBillExport(w)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	name := fmt.Sprintf("业务线订单-%d.xlsx", time.Now().Local().Unix())
	//设置请求头 使用浏览器下载
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+name)
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, bytes.NewReader(bill.Bytes()))

}

//ActionOrderItemByAdmin 业务线订单，执行扣款/拒绝（管理后台操作）
func ActionOrderItemByAdmin(c *handler.Context) error {

	req := new(domain.AccountOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	err = service.AdminVerifyBusinessOrder(user, req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

//ActionOrderItemByAccount 业务线订单，同意/拒绝（商户操作）
func ActionOrderItemByAccount(c *handler.Context) error {

	req := new(domain.AccountOperateInfo)
	err := c.Binder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	//user := c.GetTokenUser()
	err = service.AccountVerifyBusinessOrder(req)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func FindBusinessOrderBillExport(info *domain.OrderReqInfo) (bytes.Buffer, error) {
	info.Offset = 0
	info.Limit = 99999
	bill, _, err := service.SearchBusinessOrderList(info)
	if err != nil {
		return bytes.Buffer{}, err
	}
	xFile := xlsx.NewFile()
	sheet, err := xFile.AddSheet("Sheet1")
	if err != nil {
		return bytes.Buffer{}, err
	}
	title := []string{"序号", "姓名", "账户状态", "商户ID", "手机号", "邮箱",
		"交易类型", "订单ID", "收费模式", "套餐类型", "业务线ID", "业务线名称", "主链币名", "代币名",
		"部署费", "托管月费", "服务费", "押金费", "增加主链币费", "增加代币费", "优惠费", "套餐收益户",
		"扣费币种", "商户审核状态", "商户审核时间", "操作人", "订单状态", "订单时间", "备注"}
	r := sheet.AddRow()
	var ce *xlsx.Cell
	for _, v := range title {
		ce = r.AddCell()
		ce.Value = v
	}
	for i := 0; i < len(bill); i++ {
		r = sheet.AddRow()
		// 序号
		ce = r.AddCell()
		ce.Value = strconv.Itoa(i)
		// 商户名称
		ce = r.AddCell()
		ce.Value = bill[i].Name
		// 账户状态
		ce = r.AddCell()
		ce.Value = dict.IsTestText[bill[i].AccountStatus]
		// 商户ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].AccountId)
		// 手机号
		ce = r.AddCell()
		ce.Value = bill[i].Phone
		// email
		ce = r.AddCell()
		ce.Value = bill[i].Email
		// email
		ce = r.AddCell()
		ce.Value = bill[i].OrderType
		// orderId
		ce = r.AddCell()
		ce.Value = bill[i].OrderId
		ce = r.AddCell()
		ce.Value = bill[i].TypeName
		ce = r.AddCell()
		ce.Value = bill[i].ModelName
		// 业务线ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].BusinessId)
		// 业务线名
		ce = r.AddCell()
		ce.Value = bill[i].BusinessName
		// 主链币
		ce = r.AddCell()
		ce.Value = bill[i].Coin
		// 代币
		ce = r.AddCell()
		ce.Value = bill[i].SubCoin

		ce = r.AddCell()
		ce.Value = bill[i].DeployFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].CustodyFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].DepositFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].CoverFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].AddBusinessFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].AddChainFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].AddSubChainFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].DiscountFee.String()
		ce = r.AddCell()
		ce.Value = bill[i].ProfitNumber.String()
		// 扣费币种
		ce = r.AddCell()
		ce.Value = bill[i].DeductCoin
		ce = r.AddCell()
		ce.Value = service.VerifyStatus(bill[i].AccountVerifyState)
		// 订单状态
		ce = r.AddCell()
		ce.Value = bill[i].AccountVerifyTime
		// 发送地址
		ce = r.AddCell()
		ce.Value = bill[i].AdminVerifyName
		// 接收地址
		ce = r.AddCell()
		ce.Value = service.VerifyStatus(bill[i].AdminVerifyState)
		// Memo
		ce = r.AddCell()
		ce.Value = bill[i].CreateTime
		// 数量
		ce = r.AddCell()
		ce.Value = bill[i].Remark

	}
	//将数据存入buff中
	var buff bytes.Buffer
	if err = xFile.Write(&buff); err != nil {
		return bytes.Buffer{}, err
	}
	return buff, nil
}

func OffsetPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if offset <= 0 {
		offset = 0
	}

	return limit * offset, limit
}
