package controller

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"flowSign/api/models"

	//"flowSign/api/models"
	"flowSign/common/conf"
	"flowSign/common/log"
	"flowSign/common/validator"

	gosdk "github.com/okex/okexchain-go-sdk"

	//tokentypes "github.com/okex/okexchain/x/token/types"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
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

	returns.Data = "0x" + hex.EncodeToString(rawTx)
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
	log.Info(String(params))

	var returns = &validator.SignReturns{SignHeader: params.SignHeader}

	rawTx,txid,err :=this.mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	returns.Data = hex.EncodeToString(rawTx)
	returns.Txid = txid
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))
	config, err := gosdk.NewClientConfig(conf.GetConfig().Node.Url, conf.GetConfig().Node.Chainid, gosdk.BroadcastBlock, decimal.NewFromInt(int64(2000000000000000000)).Shift(-18).String()+"okt", 200000,
		0, "")
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturns)

	client := gosdk.NewClient(config)
	acc,err:=client.Auth().QueryAccount(params.Address)
	log.Info(acc.String())
	if err != nil {
		ret.Code=-1
		ret.Data = "0"
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	}
	coins :=acc.GetCoins()
	for _,coin := range coins{
		if coin.Denom == params.Token{
			ret.Code=0
			ret.Data=coin.Amount.String()
			ctx.JSON(http.StatusOK, ret)
			return
		}
	}
	ret.Code=0
	ret.Data = "0"
	ret.Data = "coin not find"
	ctx.JSON(http.StatusOK, ret)
	return

}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))
	ret := new(validator.ValidAddressReturns)
	if len(params.Address) != 48 || !strings.HasPrefix(params.Address, "okexchain") {
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
