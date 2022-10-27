package controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"net/http"
	"satSign/api/models"
	"satSign/app"
	"satSign/common/conf"
	"satSign/common/log"
	. "satSign/common/validator"
	"satSign/utils/btc"
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
	if len(addrInfo.Addrs) <= 0 {
		appG.Response(http.StatusOK, 1, errors.New("empty address"))
		return
	}
	if len(addrInfo.Addrs) > 1000 {
		appG.Response(http.StatusOK, 1, errors.New("max address is 1000"))
		return
	}
	err := importAddress(addrInfo.Addrs)
	if err != nil {
		appG.Response(http.StatusOK, 1, err.Error())
		return
	}
	appG.Response(http.StatusOK, 1, nil)
}

func importAddress(addrs []string) error {
	if len(addrs) <= 0 {
		return errors.New("empty address arr")
	}

	for _, v := range addrs {
		_, err := new(models.SatModel).ImportAddress(v, "", false)
		if err != nil {
			logrus.Error(err)
		}
	}
	return nil
}

func (this *SatController) listUnSpent(c *gin.Context) {
	appG := app.Gin{C: c}
	addrs := make([]string, 0)
	if err := c.BindJSON(&addrs); err != nil {
		this.NewError(c, err.Error())
		return
	}
	dd, _ := json.Marshal(addrs)
	logrus.Infof("SatGetListUnSpent addrs: %s", string(dd))
	if len(addrs) <= 0 {
		logrus.Info("SatGetListUnSpent len(addrs) = 0")
		this.NewError(c, "ListUnSpent len(addrs) = 0")
		return
	}
	addrsCpoy := make([]string, 0)
	for _, v := range addrs {
		if v != "" {
			addrsCpoy = append(addrsCpoy, v)
		}
	}

	info, err := new(models.SatModel).ListUnSpent(addrsCpoy)
	if err != nil {
		logrus.Errorf("SatGetListUnSpent err :%s", err.Error())
		this.NewError(c, err.Error())
		return
	}
	if info.Error != nil {
		indata, _ := json.Marshal(info)
		logrus.Errorf("SatGetListUnSpent err :%s", indata)
		appG.Response(http.StatusOK, 1, info.Error)
		return
	}
	//vo转换
	vo := make([]*btc.SatUnSpentVO, 0)
	for _, v := range info.Result {
		va := v.Amount
		va = va.Mul(decimal.NewFromFloat(100000000))
		vo = append(vo, &btc.SatUnSpentVO{
			Txid:          v.Txid,
			Vout:          v.Vout,
			Address:       v.Address,
			ScriptPubKey:  v.ScriptPubKey,
			Confirmations: v.Confirmations,
			Solvable:      v.Solvable,
			Spendable:     v.Spendable,
			Amount:        va.IntPart(),
		})
	}
	appG.Response(http.StatusOK, 0, vo)
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
	log.Info(ToJson(params))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.SatModel).SignE(params); err != nil {
		log.Info(params.OrderId, err.Error())
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
	log.Info(ToJson(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
