package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"marSign/api/models"
	"marSign/common/conf"
	"marSign/common/log"
	"marSign/common/validator"
	btc "marSign/utils/mars"
)

type Controller struct {
	mod models.EosModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/getBalance", this.getBalance)
		group.POST("/validAddress", this.validAddress)
	}
}
func (this *Controller) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
		"data":    "",
	})
}
func (this *Controller) createAddress(ctx *gin.Context) {
	var params = new(validator.CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &validator.CreateAddressReturns{
		Data: validator.CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = this.mod.NewAccount(params.Num, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(String(params))
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}

	rawTx,txid,err :=this.mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	returns.Data = rawTx
	returns.Txid = txid
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url,"","")
	log.Info(String(params))
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	rawTx,txid,err :=this.mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Data = rawTx
	returns.Txid = txid
	txid2,err :=client.SendRawTransaction(returns.Data)
	if err != nil {
		returns.Code = -1
		returns.Message= err.Error()
	}

	returns.Txid = txid2
	log.Info(String(returns),txid,txid2)
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := btc.NewRpcClient(conf.GetConfig().Node.Url,"","")
	amount,err :=client.GetBalance(params.Address)
	if err != nil{
		this.NewError(ctx,err.Error())
		return
	}
	ret:= new(validator.GetBalanceReturns)

	ret.Code=0
	ret.Data = amount
	ctx.JSON(http.StatusOK, ret)
	return

}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)
	if (len(params.Address) != 41 && len(params.Address) != 40) || !strings.HasPrefix(params.Address, "SP") {
		println(len(params.Address) ,!strings.HasPrefix(params.Address, "SP"))
		ret.Code = -1
		ret.Data = false
		ret.Message = "failed. invalid address"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
	return
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
