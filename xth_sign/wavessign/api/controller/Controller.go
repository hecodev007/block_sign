package controller

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"wavessign/api/models"

	"github.com/onethefour/common/xutils"

	//"wavessign/api/models"
	"wavessign/common/conf"
	"wavessign/common/log"
	"wavessign/common/validator"

	//tokentypes "github.com/okex/exchain-go-sdk/module/token/types"
	//tokentypes "github.com/okex/exchain/x/token/types"

	//sdk "github.com/okex/exchain-go-sdk/types"
	//sdk "github.com/okex/exchain-go-sdk/types"

	//tokentypes "github.com/okex/exchain/x/token/types"

	btc "wavessign/utils/waves"

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
		//gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}),
		group.POST("/transfer", this.transfer)
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
	//if params.Token == "okt" {
	//	params.Token = ""
	//}
	if err := xutils.LockMax(params.FromAddress, 3); err != nil {
		log.Info(fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		this.NewError(ctx, fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress, time.Second*3)

	if params.Fee == 0 {
		params.Fee = conf.GetConfig().Node.Fee
	}
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if params.Timestamp == 0 {
		params.Timestamp = uint64(time.Now().UnixMilli() + 129000)
	}
	txid, tx, err := btc.Sign(params, string(pri))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	txbytes, err := tx.MarshalBinary()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Rawtx = hex.EncodeToString(txbytes)
	returns.Data = txid
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
	if params.Fee == 0 {
		params.Fee = conf.GetConfig().Node.Fee
	}
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if params.Timestamp == 0 {
		params.Timestamp = uint64(time.Now().UnixMilli() + 129000)
	}
	if params.ContractAddress == "" {
		value, _, err := btc.GetBalance(params.FromAddress, params.ContractAddress)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if value < params.Fee+params.Value.BigInt().Uint64() {
			this.NewError(ctx, fmt.Sprintf("地址余额不足"))
			//return
		}
	} else {
		maincoin, _, err := btc.GetBalance(params.FromAddress, "")
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if maincoin < params.Fee {
			this.NewError(ctx, fmt.Sprintf("地址余额不足"))
			return
		}
		value, _, err := btc.GetBalance(params.FromAddress, params.ContractAddress)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		if value < params.Value.BigInt().Uint64() {
			this.NewError(ctx, fmt.Sprintf("地址余额不足"))
			return
		}
	}
	txid, tx, err := btc.Sign(params, string(pri))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	err = btc.SendRawTransaction(tx)
	if err != nil {
		this.NewError(ctx, "交易发送出错,联系开发确认是否上链:"+err.Error()+" txid:"+txid)
		return
	}
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	txbytes, err := tx.MarshalBinary()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Rawtx = hex.EncodeToString(txbytes)
	returns.Data = txid
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
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

	_, balance, err := btc.GetBalance(params.Address, params.Token)
	if err != nil {
		ret.Code = 1
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	} else {
		ret.Data = balance
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
	err := btc.VerifyAddress(params.Address)
	if err != nil {
		ret.Code = -1
		ret.Data = false
		ret.Message = "failed::" + err.Error()
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
