package controller

import (
	"encoding/hex"
	"encoding/json"
	"mwSign/api/models"
	"mwSign/common/conf"
	"mwSign/common/log"
	"mwSign/common/validator"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	btc "mwSign/utils/mw"

	"github.com/gin-gonic/gin"
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
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	time.Sleep(time.Second * 5)
	if params.Deadline == 0 {
		params.Deadline = 1440
	}
	if params.Fee.IntPart() == 0 {
		params.Fee = decimal.NewFromInt(100000000)
	}
	pri, err := this.Mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	pub, err := btc.PrivateToPub(string(pri))
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	buildtx, err := client.BuildTx(pub, params.ToAddress, params.Value.String(), params.Fee.String(), params.Deadline)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	txbytes, _ := hex.DecodeString(buildtx.UnsignedTransactionBytes)
	rawtx, err := btc.Sign(txbytes, string(pri))
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	returns.Data = rawtx
	returns.TxHash = ""
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))
	if params.Deadline == 0 {
		params.Deadline = 1440
	}
	if params.Fee.IntPart() == 0 {
		params.Fee = decimal.NewFromInt(100000000)
	}
	pri, err := this.Mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	pub, err := btc.PrivateToPub(string(pri))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	buildtx, err := client.BuildTx(pub, params.ToAddress, params.Value.String(), params.Fee.String(), params.Deadline)
	if err != nil {
		log.Info(String(err.Error()))
		this.NewError(ctx, err.Error())
		return
	}
	txbytes, _ := hex.DecodeString(buildtx.UnsignedTransactionBytes)
	rawtx, err := btc.Sign(txbytes, string(pri))
	if err != nil {
		log.Info(String(err.Error()))
		this.NewError(ctx, err.Error())
		return
	}
	ret, err := client.SendRawTransaction(rawtx)
	if err != nil {
		log.Info(String(err.Error()))
		this.NewError(ctx, err.Error())
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	returns.Data = rawtx
	returns.TxHash = ret.Transaction
	ctx.JSON(http.StatusOK, returns)
	log.Info(String(returns))
	return
}

func String(d interface{}) string{
	str,_ := json.Marshal(d)
	return string(str)
}