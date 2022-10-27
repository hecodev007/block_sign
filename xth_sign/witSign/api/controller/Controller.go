package controller

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"witSign/api/models"
	"witSign/common/conf"
	"witSign/common/log"
	"witSign/common/validator"

	gosdk "github.com/okex/okchain-go-sdk"

	tokentypes "github.com/okex/okchain-go-sdk/module/token/types"
	"github.com/okex/okchain-go-sdk/types"
	"github.com/okex/okchain-go-sdk/utils"

	//sdk "github.com/okex/okchain-go-sdk/types"
	sdk "github.com/okex/okchain-go-sdk/types"

	//tokentypes "github.com/okex/okexchain/x/token/types"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
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
	if params.Token == "" {
		params.Token = "okt"
	}
	log.Info(String(params))

	var returns = &validator.TransferReturns{SignHeader: params.SignHeader}
	config, err := gosdk.NewClientConfig(conf.GetConfig().Node.Url, "okexchain", gosdk.BroadcastBlock, decimal.NewFromInt(int64(params.Fee)).Shift(-4).String()+"okt", 200000,
		0, "")
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := gosdk.NewClient(config)
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	_, err = utils.CreateAccountWithPrivateKey(String(pri), params.FromAddress, "")
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	accInfo, err := client.Auth().QueryAccount(params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	accountNum, sequenceNum := accInfo.GetAccountNumber(), accInfo.GetSequence()
	tokenClient, ok := client.Token().(types.BaseClient)
	if !ok {
		this.NewError(ctx, "内部类型错误")
		return
	}
	coins, err := types.ParseDecCoins(params.Value + params.Token)
	toAddr, err := types.AccAddressFromBech32(params.ToAddress)

	fromAddr, err := types.AccAddressFromBech32(params.FromAddress)
	msg := tokentypes.NewMsgTokenSend(fromAddr, toAddr, coins)
	stdTx, err := tokenClient.BuildStdTx(params.FromAddress, "", params.Memo, []sdk.Msg{msg}, accountNum, sequenceNum)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	rawTx, err := tokenClient.GetCodec().MarshalBinaryLengthPrefixed(stdTx)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Data = "0x" + hex.EncodeToString(rawTx)
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
	if params.Token == "" {
		params.Token = "okt"
	}
	log.Info(String(params))

	var returns = &validator.TransferReturns{SignHeader: params.SignHeader}
	config, err := gosdk.NewClientConfig(conf.GetConfig().Node.Url, "okexchain", gosdk.BroadcastBlock, decimal.NewFromInt(int64(params.Fee)).Shift(-4).String()+"okt", 200000,
		0, "")
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := gosdk.NewClient(config)
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	keyinfo, err := utils.CreateAccountWithPrivateKey(String(pri), params.FromAddress, "")
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	accInfo, err := client.Auth().QueryAccount(params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	accountNum, sequenceNum := accInfo.GetAccountNumber(), accInfo.GetSequence()
	rsp, err := client.Token().Send(keyinfo, "", params.Value+params.Token, params.Memo, params.ToAddress, accountNum, sequenceNum)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	returns.Data = rsp.TxHash
	log.Info(rsp.TxHash)
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
