package controller

import (
	"fmt"
	"net/http"
	"zenSign/api/models"
	"zenSign/common"
	"zenSign/common/conf"
	"zenSign/common/log"
	btc "zenSign/utils/zcash"

	"github.com/gin-gonic/gin"
)

type ZcashController struct {
}

func (this *ZcashController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
	}
}

func (this *ZcashController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *ZcashController) createAddress(ctx *gin.Context) {

	var params = new(common.CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &common.ZcashCreateAddressReturns{
		Data: common.ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.ZcashModel).NewAccount(params.Num, params.MchId, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *ZcashController) sign(ctx *gin.Context) {
	var params = new(common.ZenSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if params.Data.BlockHash == "" {
		params.Data.BlockHash = "00000001cf4e27ce1dd8028408ed0a48edd445ba388170c9468ba0d42fff3052"
		params.Data.BlockHeight = 142091
	}
	log.Info(params.String())
	var returns = &common.ZenSignReturns{SignHeader: params.SignHeader}
	if rawTx, txhash, err := new(models.ZcashModel).SignTx(params); err != nil {
		returns.Data = rawTx
		returns.Code = -1
		returns.Message = err.Error()
	} else {
		returns.Data = rawTx
		returns.TxHash = txhash
	}
	log.Info(returns.String())
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *ZcashController) transfer(ctx *gin.Context) {
	var params = new(common.ZenSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	if params.Data.BlockHash == "" {
		log.Info("GetBlockCount")
		h, err := client.GetBlockCount()
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		h -= 300
		//h = 841324
		hash, err := client.GetBlockHash(h)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		log.Info(h)
		params.Data.BlockHash = hash
		params.Data.BlockHeight = h
		//params.Data.BlockHash = "00000001cf4e27ce1dd8028408ed0a48edd445ba388170c9468ba0d42fff3052"
		//params.Data.BlockHeight = 142091
	}
	for k, v := range params.TxIns {
		tx, err := client.GetRawTransaction(v.FromTxid)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		params.TxIns[k].FromScript = tx.Vout[v.FromIndex].ScriptPubkey.Hex
		log.Info(params.TxIns[k].FromTxid, params.TxIns[k].FromScript)
	}
	log.Info(params.String())
	var returns = &common.ZenSignReturns{SignHeader: params.SignHeader}
	if rawTx, txhash, err := new(models.ZcashModel).SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else if _, err := client.SendRawTransaction(rawTx); err != nil {
		returns.Code = -1
		returns.Message = err.Error()
		returns.TxHash = txhash
		returns.Data = rawTx
	} else {
		returns.Data = rawTx
		returns.TxHash = txhash
	}
	log.Info(returns.String())
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *ZcashController) test(ctx *gin.Context) {
	var params = new(common.ZenSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		//this.NewError(ctx, err.Error())
		//return
	}
	if wif, err := new(models.ZcashModel).GetAccount(params.MchId, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		fmt.Println("private", wif.String(), wif.PrivKey.Serialize())
		ctx.JSON(http.StatusOK, wif.String())
	}

}
