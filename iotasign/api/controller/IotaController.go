package controller

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v2"
	"github.com/sirupsen/logrus"
	"iotasign/api/models"
	"iotasign/app"
	"iotasign/common/conf"
	"iotasign/common/log"
	. "iotasign/common/validator"

	"net/http"
)

type SatController struct {
}

func (this *SatController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer" /*gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}),*/, this.transfer)
		group.POST("/unspents", this.listUnSpent)
		group.POST("/importaddrs", this.improtAddrs)
	}
}

type ImportAddrs struct {
	Addrs []string `json:"addrs"`
}

func (this *SatController) improtAddrs(c *gin.Context) {
	appG := app.Gin{C: c}
	addrInfo := &ImportAddrs{}
	if err := c.BindJSON(addrInfo); err != nil {
		this.NewError(c, err.Error())
		return
	}

	appG.Response(http.StatusOK, 1, nil)
}

func (this *SatController) listUnSpent(c *gin.Context) {
	appG := app.Gin{C: c}
	addrs := make([]string, 0)
	if err := c.BindJSON(&addrs); err != nil {
		this.NewError(c, err.Error())
		return
	}
	fmt.Println("addrs:", addrs)
	info, err := new(models.SatModel).ListUnSpent(addrs)
	if err != nil {
		logrus.Errorf("SatGetListUnSpent err :%s", err.Error())
		appG.Response(http.StatusOK, 1, err)
		this.NewError(c, err.Error())
		return
	}
	fmt.Println("ListUnSpent:", info)
	appG.Response(http.StatusOK, 0, info)
}

func (this *SatController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}

func (this *SatController) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &CreateAddressReturns{
		Data: CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.SatModel).NewAccount(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *SatController) sign(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	//pjson, _ := json.Marshal(params)
	//log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.SatModel).Sign(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *SatController) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	//log.Info(ToJson(params))
	data, _ := json.Marshal(params)
	log.Info("param:", string(data))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.SatModel).Sign(params); err != nil {
		log.Info(params.OrderId, err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx

		data, err := hex.DecodeString(rawTx)
		if err != nil {
			log.Info(params.OrderId, err)
			returns.Code = -1
			returns.Message = err.Error()
			ctx.JSON(http.StatusOK, returns)
			return
		}
		transaction := &iotago.Transaction{}
		_, err = transaction.Deserialize(data, serializer.DeSeriModeNoValidation)
		if err != nil {
			log.Info(params.OrderId, err)
			returns.Code = -1
			returns.Message = err.Error()
			ctx.JSON(http.StatusOK, returns)
			return
		}
		completeMsg := &iotago.Message{
			//Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
			Payload: transaction,
			//Nonce:   3495721389537486,
		}
		nodeAPI := iotago.NewNodeHTTPAPIClient(conf.GetConfig().Node.Url)
		mess, err := nodeAPI.SubmitMessage(context.Background(), completeMsg)
		if err != nil {
			log.Info(params.OrderId, err)
			returns.Code = -1
			returns.Message = err.Error()
			ctx.JSON(http.StatusOK, returns)
			return
		}

		returns.TxHash = iotago.MessageIDToHexString(mess.MustID())
		log.Info("txHash:", returns.TxHash)
	}
	//log.Info(ToJson(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
