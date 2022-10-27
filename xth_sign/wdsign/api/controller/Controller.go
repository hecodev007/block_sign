package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"wdsign/api/models"
	"wdsign/common/conf"
	"wdsign/common/log"
	rpc "wdsign/utils/fil"
)

type GhostController struct {
}

func (this *GhostController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
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
		params.GasFeeCap = 150000
	}
	if params.GasLimit == 0 {
		params.GasLimit = 6000000
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
	if !rpc.Limit(params.From, 20) {
		this.NewError(ctx, "limit 1 request per 20s")
		return
	}

	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var err error
	client := rpc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)

	if params.GasFeeCap == 0 {
		basefee, err := client.BaseFee()
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, "获取手续费错误,交易没上链可重推:"+err.Error())
			return
		}
		params.GasFeeCap = basefee * 3
	}

	if params.GasLimit == 0 {
		params.GasLimit = 6000000
	}
	if params.GasPremium < 100 {
		params.GasPremium = 100000
	}
	if params.GasFeeCap < params.GasPremium {
		params.GasFeeCap = 150000
	}
	if params.GasFeeCap > 50000000000000000 {
		this.NewError(ctx, "手续费过高>0.05")
		return
	}

	if params.GasFeeCap <= params.GasPremium {
		params.GasFeeCap += params.GasPremium
	}
	//params.Nonce = 371
	//params.GasPremium = 100

	if params.Nonce == 0 {
		if params.Nonce, err = client.GetNonce(params.From); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, "获取nonce错误,交易没上链可重推:"+err.Error())
			return
		}
		log.Info(params.Nonce)
	}
	pjson, _ = json.Marshal(params)
	log.Info(string(pjson))
	var balance decimal.Decimal
	//var err error
	if balance, err = client.GetBalance(params.From); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, "获取账户余额错误,交易没上链可重推:"+err.Error())
		return
	}

	sendvalue, _ := decimal.NewFromString(params.Amount)
	gasused := decimal.NewFromInt(params.GasLimit)
	gascap := decimal.NewFromInt(params.GasFeeCap)
	cost := sendvalue.Add(gasused.Mul(gascap))
	if balance.Cmp(cost) < 0 {
		log.Info()
		this.NewError(ctx, fmt.Sprintf("insufficient value:余额(%v),转账(%v)", balance.Shift(-18).String(), cost.Shift(-18).String()))
		return
	}

	var returns = &models.SignReturns{Header: params.Header}
	if signMsg, rawTx, txid, err := new(models.FilModel).SignTx(params); err != nil {
		log.Info(params.OrderNo, err.Error())
		this.NewError(ctx, "签名错误,交易没上链,可重推:"+err.Error()+" 余额:"+balance.Shift(-18).String())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid
		log.Info(params.OrderNo, txid)
		_, err := client.SendRawTransaction(signMsg)
		if err != nil {
			returns.Code = -1
			log.Info(err.Error())
			returns.Message = "交易发送错误,需要开发确认是否能重推:" + err.Error() + " 余额:" + balance.Shift(-18).String()
		} else {
			returns.TxHash = txid
		}
	}
	rets, _ := json.Marshal(returns)
	log.Info(params.OrderNo, string(rets))
	ctx.JSON(http.StatusOK, returns)
	return
}
