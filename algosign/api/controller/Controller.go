package controller

import (
	"algoSign/api/models"
	"algoSign/common/conf"
	"algoSign/common/log"
	"algoSign/common/validator"
	"encoding/hex"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/algorand/go-algorand-sdk/client/algod"
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
		//group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/transfer", this.transfer)
		//DBHFW5LEKXOD5GI3BVC5ABGWBIZTIOUBK4TYNFWCT2TPS6VCL6PFCX23Y4
		//6AKUQBK4LIREPWOKDKZ2HIKBK4YEITH2W3IDFC5PJXT3RMTXVM2AXERFLM
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
	log.Info(params.String())
	if params.TelosSignParams_Data.Fee.IntPart() == 0 {
		params.TelosSignParams_Data.Fee = decimal.NewFromInt(100)
	}
	/////
	algodClient, err := algod.MakeClient(conf.GetConfig().Node.Url, "9a5bbe7ecae7fa6d81495a78b113602d6629ddf3d5b4d540ffbea8cebdf2495c")
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	if params.TransactionParams.GenesisID == "" {
		params.TransactionParams, err = algodClient.SuggestedParams()
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
	}
	///
	txid, rawtx, err := this.Mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	returns.Data = hex.EncodeToString(rawtx)
	returns.TxHash = txid
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(params.String())
	if params.TelosSignParams_Data.Fee.IntPart() == 0 {
		params.TelosSignParams_Data.Fee = decimal.NewFromInt(1000) //最低1000起步
	}

	algodClient, err := algod.MakeClient(conf.GetConfig().Node.Url, "9a5bbe7ecae7fa6d81495a78b113602d6629ddf3d5b4d540ffbea8cebdf2495c")
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	if params.TransactionParams.GenesisID == "" {
		params.TransactionParams, err = algodClient.SuggestedParams()
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
	}

	_, rawtxbytes, err := this.Mod.SignTx(params)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	//log.Info(txid)
	resp, err := algodClient.SendRawTransaction(rawtxbytes)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	returns.Data = hex.EncodeToString(rawtxbytes)
	returns.TxHash = resp.TxID
	ctx.JSON(http.StatusOK, returns)
	log.Infof("交易成功：TxHash:%s", resp.TxID)
	return
}
