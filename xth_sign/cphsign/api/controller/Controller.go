package controller

import (
	"cphsign/api/models"
	"cphsign/common/conf"
	"cphsign/common/log"
	"cphsign/common/validator"
	"cphsign/utils"
	btc "cphsign/utils/cph"
	"encoding/json"
	"net/http"
	"strings"

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
		group.POST("/validAddress", this.validAddress)
		group.POST("/getBalance", this.GetBalance)

	}
	//r.POST("/collector",this.collector)
}
func (this *Controller) GetBalance(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	cli, err := btc.NewRpcClient(conf.GetConfig().Node.Url)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	value, err := cli.GetBalance(params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceResponse)
	ret.Data = value.String()
	ctx.JSON(http.StatusOK, ret)
	return
}

func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)
	ret.Code = -1
	ret.Data = false
	if !strings.HasPrefix(strings.ToLower(params.Address), "cph") {
		ret.Message = "签名服务校验地址前缀错误"
		return
	}

	if len(params.Address) != 43 {
		ret.Message = "签名服务校验地址长度错误"
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

	addrs, err := this.Mod.NewAccount(params.Num, params.MchName, params.OrderId)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Data.Address = addrs
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
		log.Info("limit 1 request per 30s")
		this.NewError(ctx, "limit 1 request per 30s")
		return
	}
	defer utils.Free(params.FromAddress)
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client, err := btc.NewRpcClient(conf.GetConfig().Node.Url)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}

	if params.Nonce.IsZero() {
		if nonce, err := client.PendingNonceAt(params.FromAddress); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		} else {
			params.Nonce = decimal.NewFromInt(int64(nonce))
		}
	}
	//params.Nonce = decimal.NewFromInt(int64(0))
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
		params.GasPrice = decimal.NewFromInt(18000000000)
	}
	if params.GasPrice.IsZero() {
		if price, err := client.SuggestGasPrice(); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		} else {
			params.GasPrice = decimal.NewFromBigInt(price, 0)
		}
	}
	if !params.GasPrice.Mul(params.GasLimit).LessThan(decimal.NewFromInt(1e17)) {
		log.Info("手续费>0.1")
		this.NewError(ctx, "手续费>0.1")
		return
	}
	if params.Token == "" {
		balance, err := client.GetBalance(params.FromAddress)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if params.GasPrice.Mul(params.GasLimit).Add(params.Value).Cmp(balance) > 0 {
			log.Info("insuffient balance")
			this.NewError(ctx, "insuffient balance,出账额度:"+params.GasPrice.Mul(params.GasLimit).Add(params.Value).Shift(-18).String()+"   实际额度:"+balance.Shift(-18).String())
			return
		}
	} else {
		//todo:token 余额校验
	}
	//this.NewError(ctx, params.String())
	log.Info(params.String())
	//return

	if _, rawTx, err := this.Mod.SignTx(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else if txid, err := client.SendRawTransaction(rawTx); err != nil {
		log.Info(err.Error(), txid)
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Message = "success"
		returns.Rawtx = rawTx
		returns.Data = txid
		log.Info(String(returns))
		ctx.JSON(http.StatusOK, returns)
		return
	}
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
