package controller

import (
	"cruSign/api/models"
	"cruSign/common/conf"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GhostController struct {
}

func (this *GhostController) Router(r *gin.Engine) {
	group := r.Group("/v1/"+conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.GET("/test", this.test)
	}
}

func (this *GhostController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *GhostController) createAddress(ctx *gin.Context) {
	var params = new(models.CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &models.CreateAddressReturns{
		Data: models.CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.CommonModel).NewAccount(params.Num, params.MchName, params.OrderNo); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) sign(ctx *gin.Context) {
	var params = new(models.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &models.SignReturns{Header: params.Header}
	if rawTx, txid, err := new(models.CommonModel).SignTx(params); err != nil {
		fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) transfer(ctx *gin.Context) {
	var params = new(models.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &models.SignReturns{Header: params.Header}
	if rawTx, txid, err := new(models.CommonModel).SignTx(params); err != nil {
		fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid

		//client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		//txid, err := client.SendRawTransaction(rawTx)
		//if err != nil {
		//	returns.Code = -1
		//	fmt.Println(err.Error())
		//	returns.Message = err.Error()
		//} else {
		//	returns.TxHash = txid
		//}
	}

	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *GhostController) test(ctx *gin.Context) {

}
