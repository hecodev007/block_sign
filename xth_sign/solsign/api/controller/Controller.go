package controller

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"solsign/api/models"
	"solsign/common/conf"
	"solsign/common/log"
	"solsign/common/validator"
	rpc "solsign/utils/sol"

	"github.com/onethefour/common/xutils"

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
		group.POST("/validAddress", this.validAddress)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"kfal2^*&*()()": "$%%^&dkfas"}), this.transfer)
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
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	client := rpc.NewClient(conf.GetConfig().Node.Url)
	istokenAddress, _, err := client.GetAccountInfo(params.Address)
	result := validator.ValidAddressResult{
		Code:    0,
		Message: "",
		Data:    true,
	}
	if err != nil {
		result.Code = 1
		result.Message = err.Error()
		result.Data = false
		ctx.JSON(http.StatusOK, result)
		return
	}
	if istokenAddress {
		result.Code = 1
		result.Message = "请使用主链地址"
		result.Data = false
		ctx.JSON(http.StatusOK, result)
		return
	}
	ctx.JSON(http.StatusOK, result)
	return

}
func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	if rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = hex.EncodeToString(rawtx)
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	xutils.Lock(params.FromAddress)
	defer xutils.Unlock(params.FromAddress)
	if params.FeeAddress != "" {
		xutils.Lock(params.FeeAddress)
		defer xutils.Unlock(params.FeeAddress)
	}
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client := rpc.NewClient(conf.GetConfig().Node.Url)

	if params.ContractAddress == "" {
		fromBalance, err := client.GetBalance(params.FromAddress)
		if err != nil {
			log.Info(params.OrderId, err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if params.Amount.BigInt().Uint64() > fromBalance {
			log.Info(params.OrderId, fmt.Sprintf("%v 账户余额不足:%v(额度)<%v(出账)", params.FromAddress, fromBalance, params.Amount.String()))
			this.NewError(ctx, fmt.Sprintf("%v 账户余额不足:%v(额度)<%v(出账)", params.FromAddress, fromBalance, params.Amount.String()))
			return
		}
	} else {
		fromTokenBalance, _, _, err := client.BalanceOf(params.FromAddress, params.ContractAddress)
		if err != nil {
			log.Info(params.OrderId, err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if fromTokenBalance.Cmp(params.Amount) < 0 {
			log.Info(params.OrderId, fmt.Sprintf("%v 账户余额不足:%v(额度)<%v(出账)", params.FromAddress, fromTokenBalance.String(), params.Amount.String()))
			this.NewError(ctx, fmt.Sprintf("%v 账户余额不足:%v(额度)<%v(出账)", params.FromAddress, fromTokenBalance.String(), params.Amount.String()))
			return
		}
	}
	block, err := client.GetRecentBlockhash()
	if err != nil {
		log.Info(params.OrderId, err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	params.BlockHash = block.Blockhash
	//this.NewError(ctx, params.String())
	log.Info(params.String())
	//return

	if rawTx, err := this.Mod.SignTx(params); err != nil {
		log.Info(params.OrderId, err.Error())
		this.NewError(ctx, err.Error())
		return
	} else if txid, err := client.SendRawTransaction(rawTx); err != nil {
		log.Info(params.OrderId, err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Message = "success"
		returns.Data = txid
		returns.Rawtx = hex.EncodeToString(rawTx)
		ctx.JSON(http.StatusOK, returns)
		log.Info(xutils.String(returns))
		return
	}
}
