package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"waxsign/api/models"
	"waxsign/common/conf"
	"waxsign/common/log"
	"waxsign/common/validator"

	eos "github.com/eoscanada/eos-go"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type EosController struct {
	mod models.EosModel
}

func (this *EosController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/getBalance", this.getBalance)
		group.POST("/validAddress", this.validAddress)
	}
}
func (this *EosController) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
		"data":    "",
	})
}
func (this *EosController) createAddress(ctx *gin.Context) {
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
func (this *EosController) sign(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if params.Token == "" {
		params.Token = "eosio.token"
	}
	eosApi := eos.New(conf.GetConfig().Node.Url)
	info, err := eosApi.GetInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.SignParams_Data.BlockID = info.LastIrreversibleBlockID.String()
	log.Info(String(params))
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	if pack, hash, err := this.mod.SignTx(params.MchName, params.SignParams_Data.SignPubKey, &params.SignParams_Data); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		//returns.Data.PackedTransaction
		v, ok := pack.(*eos.PackedTransaction)
		if !ok {
			this.NewError(ctx, "unexpect type eos.PackedTransaction")
			return
		}
		returns.Data.PackedTransaction = *v
		returns.Data.TxHash = hash
		log.Info(String(returns))
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *EosController) transfer(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if params.Token == "" {
		params.Token = "eosio.token"
	}
	var returns = &validator.TransferReturns{SignHeader: params.SignHeader}
	eosApi := eos.New(conf.GetConfig().Node.Url)
	info, err := eosApi.GetInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.SignParams_Data.BlockID = info.LastIrreversibleBlockID.String()
	log.Info(String(params))

	asert, err := eosApi.GetCurrencyBalance(eos.AccountName(params.FromAddress), "", eos.AccountName(params.Token))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if len(asert) == 1 {
		amount := strings.Split(asert[0].String(), " ")
		tovalue := strings.Split(params.Quantity, " ")
		if strings.ToUpper(amount[1]) == strings.ToUpper(tovalue[1]) {
			amountd, _ := decimal.NewFromString(amount[0])
			tovalued, _ := decimal.NewFromString(tovalue[0])
			if amountd.Cmp(tovalued.Add(decimal.NewFromFloat(0.3))) < 0 {
				this.NewError(ctx, "余额不足:"+amountd.String()+"<"+tovalue[0]+"+0.3")
				return
			}
		}
	}

	if pack, txhash, err := this.mod.SignTx(params.MchName, params.SignParams_Data.SignPubKey, &params.SignParams_Data); err != nil {
		log.Error(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else if txid, err := eosApi.PushTransaction(pack.(*eos.PackedTransaction)); err != nil {
		log.Error(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = txid.TransactionID
		log.Info(txhash)
		log.Info(String(returns))
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *EosController) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	eosApi := eos.New(conf.GetConfig().Node.Url)
	rsp, err := eosApi.GetCurrencyBalance(eos.AccountName(params.Address), params.Params.Symbol, eos.AccountName(params.Token))
	log.Info(String(rsp))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturns)
	if len(rsp) == 0 {
		ret.Data = "0"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	if len(rsp) > 1 {
		ret.Code = -1
		ret.Message = "需要限定params.symbol"
		ret.Data = "0"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Data = decimal.NewFromInt(int64(rsp[0].Amount)).String()
	ctx.JSON(http.StatusOK, ret)
	return

}
func (this *EosController) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	eosApi := eos.New(conf.GetConfig().Node.Url)
	ret := new(validator.ValidAddressReturns)

	_, err := eosApi.GetAccount(eos.AccountName(params.Address))
	if err != nil {
		ret.Code = -1
		ret.Data = false
		ret.Message = err.Error()
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
