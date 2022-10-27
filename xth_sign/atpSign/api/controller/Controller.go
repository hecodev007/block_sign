package controller

import (
	"atpSign/api/models"
	"atpSign/common/conf"
	"atpSign/common/log"
	"atpSign/common/validator"
	"atpSign/utils"
	node "atpSign/utils/alaya"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
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
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	if txhash, rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {

		returns.Data = rawtx
		returns.TxHash = txhash
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
	if !utils.Limit(params.FromAddress, 30) {
		this.NewError(ctx, "limit 1 request per 30s")
		return
	}
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client := node.NewRpcClient(conf.GetConfig().Node.Url, "", "")

	if params.Nonce.IsZero() {
		if nonce, err := client.PendingNonceAt(params.FromAddress); err != nil {
			this.NewError(ctx, err.Error())
			return
		} else {
			params.Nonce = decimal.NewFromInt(int64(nonce))
		}
	}
	//

	//gaslimit
	if params.GasLimit.IsZero() {
		if params.Token == "" {
			params.GasLimit = decimal.NewFromInt(21000)
		} else { //代币交易gaslimit
			params.GasLimit = decimal.NewFromInt(100000)
		}
	}
	if params.GasPrice.IsZero() {
		if price, err := client.SuggestGasPrice(); err != nil {
			this.NewError(ctx, err.Error())
			return
		} else {
			params.GasPrice = decimal.NewFromBigInt(price, 0)
		}
	}
	if params.Token == "" {
		balance, err := client.BalanceAt(params.FromAddress)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		if params.GasPrice.Mul(params.GasLimit).Add(params.Value).Cmp(decimal.NewFromBigInt(balance, 0)) > 0 {
			this.NewError(ctx, "insuffient balance,出账额度:"+params.GasPrice.Mul(params.GasLimit).Add(params.Value).Shift(-18).String()+"   实际额度:"+decimal.NewFromBigInt(balance, -18).String())
			return
		}
	} else {
		//todo:token 余额校验
	}
	//this.NewError(ctx, params.String())
	log.Info(params.String())
	//return

	if txhash, rawTx, err := this.Mod.SignTx(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else if err := this.Mod.SendRawTransaction(rawTx); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Message = "success"
		returns.Data = txhash
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
