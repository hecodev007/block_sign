package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/flynn/register-service/model"
	"github.com/group-coldwallet/flynn/register-service/services"
	"github.com/group-coldwallet/flynn/register-service/util/httpresp"
)

func DeleteWatchAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.RemoveRequest
		err error
	)

	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		message := "Parse create address post data error"
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, message)
		return
	}
	if len(req.Addresses) > 1000 || len(req.Addresses) <= 0 {

		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, fmt.Sprintf("删除地址数目超过1000或者等于0，Num=%d", len(req.Addresses)))

		return
	}
	if req.Name == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "删除参数币种名字为空")
		return
	}
	err = services.DeleteWatchAddress(&req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.DELETE_ADDRESS_ERROR, fmt.Sprintf("删除监听地址错误： %v", err))
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}

func DeleteContractAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.RemoveContractRequest
		err error
	)

	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		message := "Parse create address post data error"
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, message)
		return
	}

	if req.Name == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "删除参数token币种名字为空")
		return
	}
	if req.CoinType == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "删除参数t主链币种名字为空")
		return
	}
	if req.ContractAddress == "" {
		httpresp.HttpRespError(c, httpresp.PARAM_ERROR, "删除参数合约为空")
		return
	}
	err = services.DeleteContractAddress(&req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.DELETE_ADDRESS_ERROR, fmt.Sprintf("删除合约地址错误： %v", err))
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}
