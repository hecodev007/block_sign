package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"terrasign/api/models"
	"terrasign/common/conf"
	"terrasign/common/log"
	. "terrasign/common/validator"
	rpc "terrasign/utils/kava"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onethefour/common/xutils"
)

type Controller struct {
	Mod models.AtomModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/" + conf.GetConfig().Name)
	{
		group.POST("/genaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"kava": "s62HPmQgFNBE"}), this.transfer)
		group.GET("/validaddress", this.validAddress)
		group.GET("/getbalance", this.getBalance)
	}
	//r.POST("/genaddr", this.createAddress)
	//r.POST("/signtx", this.sign)
	//r.POST("/transfer", gin.BasicAuth(gin.Accounts{"kava": "kava18859"}), this.transfer)
	//r.GET("/validaddress", this.validAddress)

}

func (this *Controller) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}

func (this *Controller) getBalance(ctx *gin.Context) {
	Address := ctx.Query("address")
	ret := new(GetBalanceReturns)
	if len(Address) != 44 || !strings.HasPrefix(Address, "terra") {
		ret.Code = -1
		ret.Data = "0"
		ret.Message = Address + " 地址校验失败,长度!=43,或没有前缀'kava'"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	node := rpc.NewNodeClient(conf.GetConfig().Node.Node)
	amount, _, _, err := node.AuthAccount(Address)
	if err != nil {
		ret.Code = -1
		ret.Data = "0"
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = strconv.FormatInt(amount, 10)
	ctx.JSON(http.StatusOK, ret)
	return
}

func (this *Controller) validAddress(ctx *gin.Context) {
	Address := ctx.Query("address")
	ret := new(ValidAddressReturns)
	if len(Address) != 44 || !strings.HasPrefix(Address, "terra") {
		ret.Code = -1
		ret.Data = false
		ret.Message = Address + "地址校验失败,长度!=43,或没有前缀'kava'"
		ctx.JSON(http.StatusOK, ret)
		return
	}

	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
	return
}
func (this *Controller) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &CreateAddressReturns{
		Data: CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = this.Mod.NewAccount(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {

	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	//params.Amount = params.Amount_str.IntPart()
	if params.Data.Fee == 0 {
		params.Data.Fee = 2500
	}
	if params.Data.Gas == 0 {
		params.Data.Gas = 100000
	}
	if params.Data.AccountNumber == 0 || params.Data.Sequence == 0 {
		var err error
		var amount int64
		node := rpc.NewNodeClient(conf.GetConfig().Node.Node)
		num := 0
	authaccount:
		amount, params.Data.AccountNumber, params.Data.Sequence, err = node.AuthAccount(params.Data.FromAddr)
		num++
		if err != nil {
			if num < 3 {
				goto authaccount
			}
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if amount < params.Data.Amount+params.Data.Fee {
			this.NewError(ctx, fmt.Sprintf("insufficient balance:%v(balance)<%v(toamount)+%v(fee)", amount, params.Data.Amount, params.Data.Fee))
			return
		}
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := this.Mod.Sign(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.RawTx = rawTx
	}

	ctx.JSON(http.StatusOK, returns)
	log.Info(String(returns))
	return
}

func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if err := xutils.LockMax(params.Data.FromAddr, 3); err != nil {
		this.NewError(ctx, "交易频率限制")
		return
	}
	defer xutils.UnlockDelay(params.Data.FromAddr, time.Second*3)
	//params.Amount = params.Amount_str.IntPart()
	if params.Data.Fee == 0 {
		params.Data.Fee = 2500
	}
	if params.Data.Gas == 0 {
		params.Data.Gas = 100000
	}
	if params.Data.AccountNumber == 0 || params.Data.Sequence == 0 {
		var err error
		var amount int64
		node := rpc.NewNodeClient(conf.GetConfig().Node.Node)
		num := 0
	authaccount:
		amount, params.Data.AccountNumber, params.Data.Sequence, err = node.AuthAccount(params.Data.FromAddr)
		num++
		if err != nil {
			if num < 3 {
				goto authaccount
			}
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if amount < params.Data.Amount+params.Data.Fee {
			this.NewError(ctx, fmt.Sprintf("insufficient balance:%v(balance)<%v(toamount)+%v(fee)", amount, params.Data.Amount, params.Data.Fee))
			return
		}
	}

	//pjson, _ := json.Marshal(params)
	log.Info(String(params))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := this.Mod.Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.RawTx = rawTx
		client := rpc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		txid, err := client.SendRawTransaction(rawTx)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else {
			returns.Data = txid
		}
	}
	//pjson, _ = json.Marshal(returns)
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
