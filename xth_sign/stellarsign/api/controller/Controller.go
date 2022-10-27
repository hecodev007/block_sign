package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stellarsign/api/models"
	"time"

	"github.com/stellar/go/clients/horizonclient"

	"github.com/onethefour/common/xutils"

	//"stellarsign/api/models"
	"stellarsign/common/conf"
	"stellarsign/common/log"
	"stellarsign/common/validator"

	//tokentypes "github.com/okex/exchain-go-sdk/module/token/types"
	//tokentypes "github.com/okex/exchain/x/token/types"

	//sdk "github.com/okex/exchain-go-sdk/types"
	//sdk "github.com/okex/exchain-go-sdk/types"

	//tokentypes "github.com/okex/exchain/x/token/types"

	btc "stellarsign/utils/stellar"

	"github.com/gin-gonic/gin"
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
		group.POST("/trustline", this.trustline)
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

	if err := xutils.LockMax(params.FromAddress, 3); err != nil {
		log.Info(fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		this.NewError(ctx, fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress, time.Second*3)
	seed, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	params.Seed = string(seed)
	rawtx, err := btc.BuildTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	returns.Rawtx = rawtx
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
	//打印发送过来的参数
	log.Info(xutils.String(params))
	//不能并发的链,给from地址加锁,防止nonce冲突
	//max参数是等待任务超过这个数量就立刻返回不处理,防止请求超时
	if err := xutils.LockMax(params.FromAddress, 3); err != nil {
		this.NewError(ctx, params.OrderId+" "+fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	//延迟3秒解锁,因为有些链交易发送到节点后,nonce不会立马更新,依然会冲突
	defer xutils.UnlockDelay(params.FromAddress, time.Second*3)

	//默认参数
	seed, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	params.Seed = string(seed)
	rawtx, err := btc.BuildTx(params)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	cli := horizonclient.DefaultPublicNetClient
	txresp, err := cli.SubmitTransactionXDR(rawtx)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}

	//返回结果再打印次,尤其是rawtx,出问题可能需要重发这笔交易
	log.Info(xutils.String(txresp))
	returns.Data = txresp.ID
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) trustline(ctx *gin.Context) {
	var params = new(validator.TrustLineParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	seed, err := this.mod.GetPrivate(params.MchName, params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.Seed = string(seed)
	txhash, err := btc.TrustLine(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturns)
	ret.Code = 0
	ret.Data = txhash
	ret.Message = ""
	ctx.JSON(http.StatusOK, ret)

}
func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))
	ret := new(validator.GetBalanceReturns)
	ret.Code = 0
	ret.Data = "0"
	ret.Message = ""

	amount, err := btc.GetBalance(params.Address, params.Token)
	if err != nil {
		ret.Code = 1
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	} else {
		ret.Data = amount
		ctx.JSON(http.StatusOK, ret)
		return
	}
}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))
	ret := new(validator.ValidAddressReturns)
	cli := horizonclient.DefaultPublicNetClient
	_, err := cli.AccountDetail(horizonclient.AccountRequest{AccountID: params.Address})
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
