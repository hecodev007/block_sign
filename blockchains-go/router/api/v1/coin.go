package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
)

func GetCoinList(c *gin.Context) {
	result, err := api.CoinService.GetCoinList()
	if err != nil {
		log.Errorf("GetCoinList error:%s", err.Error())
		httpresp.HttpRespErrorOnly(c)
		return
	}
	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
}
