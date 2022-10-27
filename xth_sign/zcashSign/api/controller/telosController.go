package controller

import (
	"fmt"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/gin-gonic/gin"
	"net/http"
	"zcashSign/api/models"
)

type TelosController struct {
}

func (this *TelosController) Router(r *gin.Engine) {
	group := r.Group("/v1/telos")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.GET("/test", this.test)
	}
}

func (this *TelosController) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
	})
}

func (this *TelosController) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &ZcashCreateAddressReturns{
		Data: ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.TelocModel).NewAccount(params.Num, params.MchId, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *TelosController) sign(ctx *gin.Context) {
	var params = new(TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &TelosSignReturns{SignHeader: params.SignHeader}
	if pack, err := new(models.TelocModel).SignTx(params.MchId, params.Data.SignPubKey, params.Data); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = pack
		//pri, _ := new(models.TelocModel).GetPrivate(params.MchId, params.Pubkey)
		//returns.TxHash = string(pri)
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *TelosController) test(ctx *gin.Context) {

	wif := "5KD7jgnwmZrcJ1naL9HBcVBfXREA6psBg5khbNXfHEch974m8oN"
	pri, _ := ecc.NewPrivateKey(wif)
	fmt.Println(pri.PublicKey().String())
}
