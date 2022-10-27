package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"telosSign/api/models"
)

type ZcashController struct {
}

func (this *ZcashController) Router(r *gin.Engine) {
	group := r.Group("/v1/zcash")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/test", this.test)
	}
}

func (this *ZcashController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *ZcashController) createAddress(ctx *gin.Context) {

	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &ZcashCreateAddressReturns{
		Data: ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.ZcashModel).NewAccount(params.Num, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *ZcashController) sign(ctx *gin.Context) {
	var params = new(ZcashSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &ZcashSignReturns{SignHeader: params.SignHeader}
	if rawTx, err := new(models.ZcashModel).SignTx(params.MchName, params.Data); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}

	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *ZcashController) test(ctx *gin.Context) {
	var params = new(ZcashSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		//this.NewError(ctx, err.Error())
		//return
	}
	if wif, err := new(models.ZcashModel).GetAccount(params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		fmt.Println("private", wif.String(), wif.PrivKey.Serialize())
		ctx.JSON(http.StatusOK, wif.String())
	}

}
