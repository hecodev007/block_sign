package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
)

//时间查询交易详情
func FindTransactionList(c *gin.Context) {
	params := model.SearchTx{}
	c.BindJSON(&params)
	if err := params.Check(); err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL, err.Error(), "")
		return
	}
	list, total, err := api.TxinfoService.FindTxList(&params)
	if err != nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}
	data := make(map[string]interface{})
	data["total"] = total
	data["list"] = list
	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), data)
}
