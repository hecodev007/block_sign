package controller

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/base"
	"custody-merchant-admin/router/web/handler"
	"strings"
)

// SearchCoinList 主链币列表
func SearchCoinList(c *handler.Context) error {

	data, err := base.SearchChainsList()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	back := make(map[string]interface{})
	back["list"] = data
	res := handler.NewSuccessByStruct(back)

	return res.ResultOk(c)

}

// SearchSubCoinList 代币列表
func SearchSubCoinList(c *handler.Context) error {
	req := new(domain.CoinReqInfo)
	err := c.Binder(req)
	var ids []string
	if len(req.Name) > 0 {
		ids = strings.Split(req.Name, ",")
	}
	data, err := base.SearchSubCoinListByIds(ids)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	back := make(map[string]interface{})
	back["list"] = data
	res := handler.NewSuccessByStruct(back)

	return res.ResultOk(c)

}
