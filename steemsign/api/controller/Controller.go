package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"steemsign/api/models"
	"steemsign/common/conf"
	"steemsign/common/log"
	"steemsign/common/validator"
	"steemsign/utils/rpc/types"
)

type EosController struct {
}

//curl -X POST --url http://127.0.0.1:14056/steem/transfer -d '{"mch_name":"1","coin_name":"steem","from_address":"steemgoapi","to_address":"marjay","quantity":"0.002"}'

func (this *EosController) Router(r *gin.Engine) {
	group := r.Group("/" + conf.GetConfig().CoinName)
	{
		group.POST("/sign", this.sign)
		group.POST("/transfer", this.transfer)
	}
}
func (this *EosController) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": err,
	})
}

func (this *EosController) sign(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info("/sign and get params:", String(params))
	model := &models.SteemModel{
		Url:    conf.GetConfig().Url,
		JsFile: "./offline_signing/sign_offline_transfer.js",
	}
	trans, err := model.SignTx(params.SignParams_Data)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		fmt.Println("SignTx:", err)
		return
	}
	result := new(validator.SignReturns)
	result.Data.Transaction = trans.(*types.Transaction)
	result.Data.Signatures = result.Data.Transaction.Signatures
	ctx.JSON(http.StatusOK, result)
	log.Info("return success:", String(result))
	return
}

func (this *EosController) transfer(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		fmt.Println("should bind json err")
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info("req data:", params)
	model := &models.SteemModel{
		Url:    conf.GetConfig().Url,
		JsFile: "./offline_signing/sign_offline_transfer.js",
	}
	err, txid := model.Transfer(params.SignParams_Data)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		fmt.Println("transfer err:", err)
		return
	}
	result := new(validator.TransferReturns)
	//result.Data.TxHash = txid
	result.Data = txid
	result.Message = "ok"
	ctx.JSON(http.StatusOK, result)
	return

}
func String(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}
