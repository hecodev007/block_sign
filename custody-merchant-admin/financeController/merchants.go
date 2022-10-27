package financeController

import (
	conf "custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"fmt"
)

// FindFinanceAssetsList
// 财务资产查询接口
func FindFinanceAssetsList(c *handler.Context) error {
	req := new(domain.GetAddrInfo)
	as := new(domain.AssetsSelect)
	req.Chain = c.QueryParam("chain")
	req.Coin = c.QueryParam("coin")
	as.Offset, as.Limit = c.OffsetPage()
	// 查询币种
	if req.Chain != "" && req.Coin != "" {
		coin, err := base.FindCoinsByChainName(req.Coin, req.Chain)
		if err != nil {
			return handler.OutCodeError(c, 30006, err.Error())
		}
		if coin.Id == 0 {
			return handler.OutCodeError(c, 30006, "主链币的币种暂无")
		}
		as.CoinId = int(coin.Id)
	}
	as.CoinName = req.Coin
	dataMap, err := service.FindFinanceAssetsList(as)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	return handler.OutResult(c, 10000, "success", dataMap)
}

func PassOrRefuseMerchant(c *handler.Context) error {

	return handler.OutResult(c, 10000, "success", map[string]interface{}{})
}

func FreezeMerchantAssets(c *handler.Context) error {

	return handler.OutResult(c, 10000, "success", map[string]interface{}{})
}

func FreezeMerchantAccount(c *handler.Context) error {

	return handler.OutResult(c, 10000, "success", map[string]interface{}{})
}

func PushMerchant(c *handler.Context) error {

	xkutils.PostJson(fmt.Sprintf("%s", conf.Conf.Finance.Url), map[string]interface{}{})
	return handler.OutResult(c, 10000, "success", map[string]interface{}{})
}
