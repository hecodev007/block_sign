package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/adminPermission/auditRole"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"strings"
	"time"
)

// FindPhoneCodeAll
// 基础查询
// 获取手机区域列表
func FindPhoneCodeAll(c *handler.Context) error {
	var (
		pList []domain.PhoneInfo
		err   error
	)
	cache.GetRedisClientConn().Get(global.PhoneCode, &pList)
	if len(pList) == 0 {
		pList, err = service.FindPhoneCodeAllService()
		if err != nil {
			return handler.NewError(c, err.Error())
		}
		cache.GetRedisClientConn().Set(global.PhoneCode, pList, 8*time.Hour)
	}
	res := handler.NewSuccess()
	res.AddData("list", pList)
	return res.ResultOk(c)
}

func FindMerchantHaveAllService(c *handler.Context) error {
	id := xkutils.StrToInt64(c.QueryParam("merchant_id"))
	audit, err := service.GetUserServiceAudit(id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", audit)
	return res.ResultOk(c)
}

func FindAllService(c *handler.Context) error {
	id := xkutils.StrToInt64(c.QueryParam("merchant_id"))
	audit, err := service.GetAllMerchantService(id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", audit)
	return res.ResultOk(c)
}

func FindAllUnit(c *handler.Context) error {
	unit, err := service.FindAllUnit()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", unit)
	return res.ResultOk(c)
}

func GetBaseSysRoles(c *handler.Context) error {
	unit, err := service.GetSysRoleAll(c.GetTokenUser().Role)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", unit)
	return res.ResultOk(c)
}

func GetBaseMerchantRoles(c *handler.Context) error {
	lst, err := service.GetMerchantRoleAll()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", lst)
	return res.ResultOk(c)
}

func FindAllAudit(c *handler.Context) error {
	all, err := auditRole.GetAuditLevelAll()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", all)
	return res.ResultOk(c)
}

func FindAllChain(c *handler.Context) error {
	all, err := base.FindAllChainCoins()
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", all)
	return res.ResultOk(c)
}

func FindAuditStateList(c *handler.Context) error {

	res := handler.NewSuccess()
	mpList := []map[string]interface{}{}
	mpList = append(mpList, map[string]interface{}{
		"id":   -1,
		"name": "全部",
	})
	for i, _ := range dict.AuditTypeList {
		mpList = append(mpList, map[string]interface{}{
			"id":   i,
			"name": dict.AuditTypeList[i],
		})
	}
	res.AddData("list", mpList)
	return res.ResultOk(c)
}

func FindAuditResultList(c *handler.Context) error {
	res := handler.NewSuccess()
	mpList := []map[string]interface{}{}
	mpList = append(mpList, map[string]interface{}{
		"id":   -1,
		"name": "全部",
	})
	for i, _ := range dict.OrderResult {
		mpList = append(mpList, map[string]interface{}{
			"id":   i,
			"name": dict.OrderResult[i],
		})
	}
	res.AddData("list", mpList)
	return res.ResultOk(c)
}

func FindBillTxTypeList(c *handler.Context) error {
	res := handler.NewSuccess()
	mpList := []map[string]interface{}{}
	mpList = append(mpList, map[string]interface{}{
		"id":   -1,
		"name": "全部",
	})
	for i, _ := range dict.BillState {
		mpList = append(mpList, map[string]interface{}{
			"id":   i,
			"name": dict.BillState[i],
		})
	}
	res.AddData("list", mpList)
	return res.ResultOk(c)
}

func FindMidService(c *handler.Context) error {
	mId := c.SwitchType("mid", "int64").(int64)
	res := handler.NewSuccess()
	if mId == 0 {
		res.AddData("list", []domain.DictInfo{})
	}
	chain, err := service.FindServiceCoinChain(mId, 0)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	lst := []domain.DictInfo{}
	for _, entity := range chain {
		lst = append(lst, domain.DictInfo{
			Id:   int(entity.Id),
			Name: entity.Name,
		})
	}
	res.AddData("list", lst)
	return res.ResultOk(c)
}

func FindSidCoinList(c *handler.Context) error {
	sId := c.SwitchType("sid", "int").(int)
	res := handler.NewSuccess()
	if sId == 0 {
		res.AddData("chainList", []domain.DictInfo{})
		res.AddData("coinList", []domain.DictInfo{})
	}
	info, err := service.FirstServiceCoinChain(0, sId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	coinLst := []domain.DictInfo{}
	chainLst := []domain.DictInfo{}
	chains := strings.Split(info.Coin, ",")
	subCoins := strings.Split(info.SubCoin, ",")

	ch, err := base.FindChainsInName(chains)
	sub, err := base.FindCoinsInName(subCoins)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	for _, entity := range ch {
		chainLst = append(chainLst, domain.DictInfo{
			Id:   entity.Id,
			Name: entity.Name,
		})
	}
	for _, entity := range sub {
		coinLst = append(coinLst, domain.DictInfo{
			Id:   int(entity.Id),
			Name: entity.Name,
		})
	}
	res.AddData("chainList", chainLst)
	res.AddData("coinList", coinLst)
	return res.ResultOk(c)
}
