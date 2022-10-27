package controller

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"stxSign/api/models"
	"stxSign/common/conf"
	"stxSign/common/log"
	"stxSign/common/validator"
	"stxSign/utils"
	btc "stxSign/utils/stx"
	"github.com/onethefour/common/xutils"
	"time"
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
	log.Info(err)
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
	client := btc.NewRpcClient(conf.GetConfig().Node.Url,"","")

	if params.Fee.IsZero() {
		params.Fee = decimal.NewFromInt(conf.GetConfig().Node.Fee).Shift(-6)
	}

	if params.Nonce == 0{
		if nonce,err := client.GetNonce(params.FromAddress);err != nil{
			this.NewError(ctx,err.Error())
			return
		} else {
			params.Nonce = nonce
		}
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
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(xutils.String(params))
	if err := xutils.LockMax(params.FromAddress,3);err != nil {
		this.NewError(ctx, params.OrderId+" "+fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress,time.Second*3)
	//if !utils.Limit(params.FromAddress,4){
	//	this.NewError(ctx, "账户交易频率限制4s")
	//	return
	//}

	client := btc.NewRpcClient(conf.GetConfig().Node.Url,"","")

	if params.Fee.IsZero() {
		params.Fee = decimal.NewFromInt(conf.GetConfig().Node.Fee).Shift(-6)
	}
	//fee!>0.1
	if params.Fee.Shift(2).IntPart()>=1{
		this.NewError(ctx, "tx.fee限制必须小于0.1")
		return
	}

	if params.Nonce == 0{
		//缓存nonce
		cachenonce := utils.Get(params.FromAddress)
		if cachenonce == 0 {
			if nonce, err := client.GetNonce(params.FromAddress); err != nil {
				this.NewError(ctx, err.Error())
				return
			} else {
				params.Nonce = nonce
				utils.Set(params.FromAddress,nonce)
			}
		}else {
			params.Nonce=cachenonce
		}
	}
	nonce,balance,err := client.GetNonceAndBalance(params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if params.Nonce < nonce{
		params.Nonce = nonce
	}
	if params.Nonce > nonce{
		var totalSended uint64
		for i:=nonce;i<params.Nonce;i++{
			sendedValue:=utils.Get(fmt.Sprintf("%v%v",params.FromAddress,i))
			totalSended+=sendedValue
		}
		//totalSended+=
		curSend := uint64(params.Value.Add(params.Fee).Shift(6).IntPart())
		if totalSended+curSend>balance{
			this.NewError(ctx,params.OrderId+" "+fmt.Sprintf("链上账户额度不够出账:交易额度(%v)+当前交易额度(%v)>链上额度(%v)",float64(totalSended)/1e6,float64(curSend)/1e6,float64(balance)/1e6))
			return
		}

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
	//returns.Data = "00000000000400698bb5edac46aa1b6ff31517ba400c0c8f3f9825000000000000000b00000000000000b40100b5f0c0fbdd387a6afc3468e0ab1dfe9de53bf59fb1b4d7c9c4c5421be0bb4b7f0bb5c61f26bf1aec792d5c0f633c45ba99760f222a943a8db97efe7c79e429c7030100000000000501ffffffffffffffffffffffffffffffffffffffff000000000000007b00000000000000000000000000000000000000000000000000000000000000000000"
	txid2,err :=client.SendRawTransaction(returns.Data)
	if err != nil {
		returns.Code = -1
		returns.Message= err.Error()
	} else {
		utils.Add(params.FromAddress,1)
		utils.Set(fmt.Sprintf("%v%v",params.FromAddress,params.Nonce),uint64(params.Value.Add(params.Fee).Shift(6).IntPart()))
	}

	returns.Txid = txid2
	log.Info(String(returns),txid,txid2)
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := btc.NewRpcClient(conf.GetConfig().Node.Url,"","")
	amount,err :=client.GetBalance(params.Address)
	if err != nil{
		this.NewError(ctx,err.Error())
		return
	}
	ret:= new(validator.GetBalanceReturns)

	ret.Code=0
	ret.Data = amount.String()
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
	if (len(params.Address) != 41 && len(params.Address) != 40) || !strings.HasPrefix(params.Address, "SP") {
		println(len(params.Address) ,!strings.HasPrefix(params.Address, "SP"))
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
