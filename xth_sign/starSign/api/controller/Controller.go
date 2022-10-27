package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"starSign/api/models"
	"starSign/common/conf"
	"starSign/common/log"
	rpc "starSign/utils/fil"
)

type GhostController struct {
}

func (this *GhostController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		//group.POST("/transfer", this.transfer)
		group.GET("/test", this.test)
	}
}

func (this *GhostController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *GhostController) createAddress(ctx *gin.Context) {
	var params = new(models.CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &models.CreateAddressReturns{
		Data: models.CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.FilModel).NewAccount(params.Num, params.MchName, params.OrderNo); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) sign(ctx *gin.Context) {
	var params = new(models.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	if params.GasFeeCap == 0 {
		params.GasFeeCap = 10000000
	}
	if params.GasLimit == 0 {
		params.GasLimit = 4000000
	}
	if params.GasPremium == 0 {
		params.GasPremium = 200000
	}
	var returns = &models.SignReturns{Header: params.Header}
	if _, rawTx, txid, err := new(models.FilModel).SignTx(params); err != nil {
		fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) transfer(ctx *gin.Context) {
	var params = new(models.SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if !rpc.Limit(params.From, 30) {
		this.NewError(ctx, "limit 1 request per 30s")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var err error
	client := rpc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)

	if params.GasFeeCap == 0 {
		basefee, _ := client.BaseFee()
		params.GasFeeCap = basefee * 2
	}
	if params.GasLimit == 0 {
		params.GasLimit = 4000000
	}
	if params.GasPremium == 0 {
		params.GasPremium = 200000
	}
	if params.GasFeeCap < params.GasPremium {
		params.GasFeeCap = 10000000
	}
	pjson, _ = json.Marshal(params)
	log.Info(string(pjson))
	if params.Nonce == 0 {
		if params.Nonce, err = client.GetNonce(params.From); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		log.Info(params.Nonce)
	}
	if balance, err := client.GetBalance(params.From); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		sendvalue, _ := decimal.NewFromString(params.Amount)
		gasused := decimal.NewFromInt(params.GasLimit)
		gascap := decimal.NewFromInt(params.GasFeeCap)
		if balance.Cmp(sendvalue.Add(gasused.Mul(gascap))) < 0 {
			this.NewError(ctx, "insufficient value")
			return
		}
	}
	var returns = &models.SignReturns{Header: params.Header}
	if signMsg, rawTx, txid, err := new(models.FilModel).SignTx(params); err != nil {
		fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid

		txid, err := client.SendRawTransaction(signMsg)
		if err != nil {
			returns.Code = -1
			fmt.Println(err.Error())
			returns.Message = err.Error()
		} else {
			returns.TxHash = txid
		}
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) test(ctx *gin.Context) {

}
