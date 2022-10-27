package controller

import (
	"encoding/json"
	"net/http"
	"satSign/api/models"
	"satSign/common/conf"
	"satSign/common/log"
	. "satSign/common/validator"
	"satSign/utils/btc"

	"github.com/gin-gonic/gin"
)

type AvaxController struct {
}

func (this *AvaxController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
	}
}

func (this *AvaxController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *AvaxController) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &CreateAddressReturns{
		Data: CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.BiwModel).NewAccount(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *AvaxController) sign(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(ToJson(params))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.BiwModel).Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *AvaxController) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(ToJson(params))

	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.BiwModel).Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		txid, err := client.SendRawTransaction(rawTx)
		if err != nil {
			log.Info(params.OrderId, err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else {
			returns.TxHash = txid
		}
	}
	log.Info(params.OrderId, ToJson(returns))

	ctx.JSON(http.StatusOK, returns)
	return
}

func ToJson(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
