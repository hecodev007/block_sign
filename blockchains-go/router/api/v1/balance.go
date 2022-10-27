package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/shopspring/decimal"
)

func GetMchBalance(c *gin.Context) {
	var balance decimal.Decimal

	params := model.MchSearchBalance{}
	c.BindJSON(&params)
	mchInfo, err := api.MchService.GetAppId(params.Sfrom)
	if err != nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}

	if params.IsContract() {
		if err = params.CheckContract(); err != nil {
			httpresp.HttpRespError(c, httpresp.FAIL, err.Error(), nil)
			return
		}

		balance, err = api.BalanceService.GetMchTokenBalance(params.TokenName, params.ContractAddress, mchInfo.Id)
	} else {
		balance, err = api.BalanceService.GetMchBalance(params.CoinName, mchInfo.Id)
	}
	if err != nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}
	mapData := make(map[string]interface{}, 0)
	mapData["balance"] = balance.String()
	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), mapData)
}
