package controller

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/router/web/handler"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

// GetIncomeList
// 收益户
// 收益户列表
func GetIncomeList(c *handler.Context) error {
	s := new(domain.SearchIncome)
	s.Account = c.QueryParam("account")
	s.MerchantId = c.SwitchType("merchant_id", "int64").(int64)
	s.ServiceId = c.SwitchType("service_id", "int").(int)
	s.ChainId = c.SwitchType("chain_id", "int").(int)
	s.StartTime = c.QueryParam("start_time")
	s.EndTime = c.QueryParam("end_time")
	s.Offset, s.Limit = c.OffsetPage()
	List, total, err := service.FindIncomePage(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", List.List)
	res.AddData("total", total)
	return res.ResultOk(c)
}

// IncomeChartInfo
// 收益户
// 收益户列表
func IncomeChartInfo(c *handler.Context) error {
	s := new(domain.SearchIncome)
	s.Account = c.QueryParam("account")
	s.MerchantId = c.SwitchType("merchant_id", "int64").(int64)
	s.ServiceId = c.SwitchType("service_id", "int").(int)
	s.CoinId = c.SwitchType("coin_id", "int").(int)
	s.StartTime = c.QueryParam("start_time")
	s.EndTime = c.QueryParam("end_time")
	s.Offset, s.Limit = c.OffsetPage()
	List, err := service.FindIncomeChart(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	ring := domain.RingObj{}
	incomeState := []domain.IncomeStateList{}
	for i, _ := range List.TopList {
		ring.RingInfo = append(ring.RingInfo, domain.RingInfo{
			Value: List.TopList[i].Price.Round(6),
			Name:  List.TopList[i].CoinName,
			ItemStyle: domain.ItemStyle{
				Color: dict.Colors[i],
			},
		})
		ring.RingItems = append(ring.RingItems, domain.RingItems{
			Value: List.TopList[i].Price.Round(6),
			Name:  List.TopList[i].CoinName,
			Color: dict.Colors[i],
		})
	}
	incomeState = append(incomeState, domain.IncomeStateList{
		IncomeName:  "总收益",
		IncomePrice: List.Totals.Round(6),
		Color:       "#51A374",
	})
	incomeState = append(incomeState, domain.IncomeStateList{
		IncomeName:  "充值收益",
		IncomePrice: List.TopUp.Round(6),
		Color:       "#FFB03A",
	})
	incomeState = append(incomeState, domain.IncomeStateList{
		IncomeName:  "提现收益",
		IncomePrice: List.Withdraw.Round(6),
		Color:       "#36C3FC",
	})
	incomeState = append(incomeState, domain.IncomeStateList{
		IncomeName:  "套餐收益",
		IncomePrice: List.Combo.Round(6),
		Color:       "#5B76F9",
	})

	res.AddData("incomeInfo", incomeState)
	res.AddData("ring", ring)
	return res.ResultOk(c)
}

// ExportIncomeList
// 收益户
// 导出收益户列表
func ExportIncomeList(c *handler.Context) error {
	s := new(domain.SearchIncome)
	s.Account = c.QueryParam("account")
	s.MerchantId = c.SwitchType("merchantId", "int64").(int64)
	s.ServiceId = c.SwitchType("serviceId", "int").(int)
	s.CoinId = c.SwitchType("coinId", "int").(int)
	s.StartTime = c.QueryParam("startTime")
	s.EndTime = c.QueryParam("endTime")
	s.Offset, s.Limit = c.OffsetPage()
	export, err := service.FindIncomeExcelExport(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	name := fmt.Sprintf("益户列表-%d.xlsx", time.Now().Unix())
	// 设置请求头 使用浏览器下载
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+name)
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, bytes.NewReader(export.Bytes()))
}
