package controller

import (
	"net/http"
	"nyzoSign/api/models"
	"nyzoSign/common/conf"
	"nyzoSign/common/validator"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"

	util "nyzoSign/utils/nyzo"
)

type Controller struct {
	Mod models.DagModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/validAddress", this.validAddress)
		group.POST("/getBalance", this.getBalance)

	}
}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)

	ok, msg := this.Mod.ValidAddress(params.Address)

	if !ok {
		ret.Code = -1
		ret.Data = false
		ret.Message = msg
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
	return
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

	var returns = &validator.ZcashCreateAddressReturns{
		Data: validator.ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = this.Mod.NewAccount(params.Num, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}

	client := util.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	pri, err := this.Mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	amount := uint64(params.Value.Shift(6).IntPart())
	if amount == 0 {
		this.NewError(ctx, "value == 0")
		return
	}
	Signature, rawtx, err := client.SendTransaction(params.FromAddress, params.ToAddress, amount, params.Memo, string(pri), false)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}

	returns.Data = rawtx
	returns.TxHash = Signature
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if len(params.Memo) > 32 {
		this.NewError(ctx, "memo字符长度超过32")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}

	client := util.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	pri, err := this.Mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	amount := uint64(params.Value.Shift(6).IntPart())
	if amount == 0 {
		this.NewError(ctx, "value == 0")
		return
	}
	Signature, rawtx, err := client.SendTransaction(params.FromAddress, params.ToAddress, amount, params.Memo, string(pri), true)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}

	returns.Data = rawtx
	returns.TxHash = Signature
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := util.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	amount, err := client.GetBalance(params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturns)

	ret.Code = 0
	ret.Data = decimal.New(int64(amount), -6).String()
	ctx.JSON(http.StatusOK, ret)
	return

}
