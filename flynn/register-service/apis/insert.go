package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/flynn/register-service/model"
	"github.com/group-coldwallet/flynn/register-service/services"
	"github.com/group-coldwallet/flynn/register-service/util/httpresp"
)

func InsertWatchAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.InsertAddressReq

		err error
	)

	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		message := "Parse create address post data error"
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, message)
		return
	}

	if len(req.Addresses) > 1000 || len(req.Addresses) <= 0 {

		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, fmt.Sprintf("插入地址数目超过1000或者等于0，Num=%d", len(req.Addresses)))

		return
	}
	if req.Name == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "插入参数币种名字为空")
		return
	}
	if req.Url == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "插入参数Url为空")
		return
	}
	err = services.InsertWatchAddress(&req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.INSERT_ADDRESS_ERROR, fmt.Sprintf("插入地址错误：%v", err))

		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}

func InsertContractInfo(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.InsertContractReq
		err error
	)

	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "Parse create address post data error")
		return
	}
	if req.Name == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "插入参数合约名字为空")
		return
	}
	if req.CoinType == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "插入参数主链名字为空")

		return
	}
	if req.ContractAddress == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "插入参数合约地址为空")

		return
	}
	err = services.InsertContractAddress(&req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.INSERT_ADDRESS_ERROR, fmt.Sprintf("插入合约地址错误：%v", err))
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}
