package controller

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
)

// GetMerchantChainList
// 链路管理
// 获取链路列表
func GetMerchantChainList(c *handler.Context) error {
	s := new(domain.SearchChains)
	s.Account = c.QueryParam("account")
	s.MerchantId = c.SwitchType("merchant_id", "int64").(int64)
	s.ServiceId = c.SwitchType("service_id", "int").(int)
	s.Offset, s.Limit = c.OffsetPage()

	chainList, total, err := service.GetMerchantChainList(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", chainList)
	res.AddData("total", total)
	return res.ResultOk(c)
}

func GetMerchantChainInfo(c *handler.Context) error {
	s := new(domain.SearchChains)
	s.Id = c.SwitchType("id", "int64").(int64)
	info, err := service.GetMerchantChainsInfo(s.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	isList := []int{}
	if info != nil && info.Id != 0 {
		if info.IsGetAddr != 0 {
			isList = append(isList, 1)
		}
		if info.IsWithdrawal != 0 {
			isList = append(isList, 2)
		}
	}
	res.AddData("result", map[string]interface{}{
		"id":      info.Id,
		"is_list": isList,
	})
	return res.ResultOk(c)
}
func SaveMerchantChainInfo(c *handler.Context) error {
	s := new(domain.UpdateChains)
	err := c.DefaultBinder(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if s.ServiceId == 0 || s.MerchantId == 0 || s.CoinId == 0 {
		return handler.NewError(c, "参数有误")
	}
	coin, err := base.FindCoinsById(s.CoinId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if coin == nil || coin.Id == 0 {
		return handler.NewError(c, "参数有误")
	}
	err = service.GetMerchantChainsByAddr(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if len(s.IsList) > 0 {
		for _, lst := range s.IsList {
			if lst == 1 {
				s.IsGetAddr = 1
			}
			if lst == 2 {
				s.IsWithdrawal = 1
			}
		}
	} else {
		s.IsGetAddr = 0
		s.IsWithdrawal = 0
	}
	// 调用钱包地址生成接口
	addrList, err := service.CreateBatchChainAddress(int64(s.ServiceId), coin.Name, 1)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	s.ChainAddr = addrList[0]
	err = service.SaveMerchantChains(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func UpdateMerchantChainInfo(c *handler.Context) error {
	s := new(domain.UpdateChains)
	var mp = map[string]interface{}{}
	err := c.DefaultBinder(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if s.Id == 0 {
		return handler.NewError(c, "参数有误")
	}
	if len(s.IsList) > 0 {
		for _, lst := range s.IsList {
			if lst == 1 {
				mp["is_get_addr"] = 1
			}
			if lst == 2 {
				mp["is_withdrawal"] = 1
			}
		}
	} else {
		mp["is_get_addr"] = 0
		mp["is_withdrawal"] = 0
	}
	err = service.UpdateMerchantChainsInfo(s.Id, mp)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func FreezeOrThawMerchantChainInfo(c *handler.Context) error {
	s := new(domain.UpdateChains)
	var mp = map[string]interface{}{}
	err := c.DefaultBinder(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	info, err := service.GetMerchantChainsInfo(s.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	mp["state"] = xkutils.ThreeDo(info.ChainState == 1, 0, 1).(int)
	mp["reason"] = s.Reason
	err = service.UpdateMerchantChainsInfo(s.Id, mp)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

func DelMerchantChainInfo(c *handler.Context) error {
	s := new(domain.SearchChains)
	err := c.DefaultBinder(s)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = service.DeleteMerchantChainsInfo(s.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}
