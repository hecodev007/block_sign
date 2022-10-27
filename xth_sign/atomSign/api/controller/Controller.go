package controller

import (
	"atomSign/api/models"
	"atomSign/common/conf"
	"atomSign/common/log"
	. "atomSign/common/validator"
	rpc "atomSign/utils/atom2"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onethefour/common/xutils"
)

type Controller struct {
	Mod models.AtomModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
	}
}

func (this *Controller) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
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
	params.Amount = params.Amount_str.IntPart()
	if params.Fee == 0 {
		params.Fee = 2500
	}
	if params.Gas == 0 {
		params.Gas = 100000
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := this.Mod.Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	if err := xutils.LockMax(params.FromAddr, 3); err != nil {
		log.Info(fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddr))
		this.NewError(ctx, fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddr))
		return
	}
	defer xutils.UnlockDelay(params.FromAddr, time.Second*5)
	//if !util.Limit(params.FromAddr, 30) {
	//	this.NewError(ctx, "limit 1 request per 20s")
	//	return
	//}
	params.Amount = params.Amount_str.IntPart()
	if params.Fee == 0 {
		params.Fee = 2500
	}
	if params.Gas == 0 {
		params.Gas = 100000
	}
	if params.AccountNumber == 0 || params.Sequence == 0 {
		var err error
		var amount int64
		node := rpc.NewNodeClient(conf.GetConfig().Node.Node)
		num := 0
	authaccount:
		amount, params.AccountNumber, params.Sequence, err = node.AuthAccount(params.FromAddr)
		num++
		if err != nil {
			if num < 3 {
				goto authaccount
			}
			this.NewError(ctx, err.Error())
			return
		}
		if amount < params.Amount+params.Fee {
			this.NewError(ctx, fmt.Sprintf("insufficient balance:%v(balance)<%v(toamount)+%v(fee)", amount, params.Amount, params.Fee))
			return
		}
	}

	pjson, _ = json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := this.Mod.Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		client := rpc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		txid, err := client.SendRawTransaction(rawTx)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else {
			returns.TxHash = txid
		}
	}
	pjson, _ = json.Marshal(returns)
	log.Info(string(pjson))
	ctx.JSON(http.StatusOK, returns)
	return
}
