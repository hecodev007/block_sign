package controller

import (
	"btmSign/api/middleware"
	"btmSign/api/models"
	"btmSign/common/conf"
	"btmSign/common/log"
	. "btmSign/common/validator"
	"btmSign/net"
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AvaxController struct {
}

func (this *AvaxController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", middleware.MaxAllowed(1) /*gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}),*/, this.transfer)
		group.POST("/getBalance", this.getBalance)
	}
}

func (this *AvaxController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}

func (this *AvaxController) getBalance(ctx *gin.Context) {
	var params = new(ReqGetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	url := conf.GetConfig().Node.Url
	id, _, err := new(models.BiwModel).GetAccountByAddress(params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	listBalancesRequest := net.ListBalancesRequest{
		AccountID: id,
	}
	listBalancesResult, err := net.Post(url+net.ListBalances, listBalancesRequest)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var lbr net.ListBalancesResult
	err = json.Unmarshal([]byte(listBalancesResult), &lbr)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if lbr.Status != "success" {
		this.NewError(ctx, listBalancesResult)
		return
	}
	var balance int64
	for _, b := range lbr.Data {
		if b.AssetID == "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" &&
			b.AccountID == id {
			balance = b.Amount
		}
	}
	fromInt := decimal.NewFromInt(balance)
	var returns = &GetBalanceResp{
		Code: 0,
		Data: fromInt.String(),
	}
	ctx.JSON(http.StatusOK, returns)
	return
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
	if returns.Data.Address, err = new(models.BiwModel).NewAccountBtm(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *AvaxController) sign(ctx *gin.Context) {
	var params = new(BtmSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.BiwModel).SignBtm2(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *AvaxController) transfer(ctx *gin.Context) {
	var params = new(BtmSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info("param: ", string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.BiwModel).SignBtm2(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		submitTransactionRequest := net.SubmitTransactionRequest{
			RawTransaction: rawTx,
		}
		submitTransactionResult, err := net.Post(conf.GetConfig().Node.Url+net.SubmitTransaction, submitTransactionRequest)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
			goto END
		}

		var str net.SubmitTransactionResult
		err = json.Unmarshal([]byte(submitTransactionResult), &str)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
			goto END
		}
		if str.Status != "success" {
			returns.Code = -1
			returns.Message = submitTransactionResult
		} else {
			returns.TxHash = str.Data.TxID
		}
	}
END:
	ctx.JSON(http.StatusOK, returns)
	return
}
