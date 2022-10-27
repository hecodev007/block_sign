package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"neoSign/api/models"
	"neoSign/common/conf"
	"neoSign/common/log"
	. "neoSign/common/validator"
	"neoSign/utils"
	"neoSign/utils/btc"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

type Controller struct {
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/get_utxos", this.getUtxos)
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
	if err := ctx.ShouldBind(params); err != nil {
		this.NewError(ctx, Error(err))
		return
	}

	var returns = &CreateAddressReturns{
		Data: CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.NeoModel).NewAccount(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, Error(err))
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, txid, err := new(models.NeoModel).Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txid
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, Error(err))
		return
	}
	for _,v := range params.TxIns{
		if !utils.Limit(fmt.Sprintf("%v_%v",v.FromTxid,v.FromIndex),30){
			this.NewError(ctx, fmt.Sprintf("%v_%v",v.FromTxid,v.FromIndex)+" limit 1 request per 30s")
			return
		}
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, txid, err := new(models.NeoModel).Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		success, err := client.NeoSendRawTransaction(rawTx)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else if !success {
			returns.Code = -1
			returns.Message = "unkown error"
		} else {
			if !strings.HasPrefix(txid,"0x"){
				txid = "0x"+txid
			}
			returns.TxHash = txid
		}
	}
	ctx.JSON(http.StatusOK, returns)
	return
}
//
func (this *Controller) getUtxos(ctx *gin.Context) {
	var params = new(GetUtxos)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, Error(err))
		return
	}
	resp, err := http.Get("https://api.neoscan.io/api/main_net/v1/get_balance/" + params.Addr)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(GetUtxosReturn)
	if err := json.Unmarshal(body, ret); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(string(body))
	ret.Address = params.Addr
	Balance := make([]*Balance, 0)
	for i, balance := range ret.Balances {
		if balance.Asset_symbol == params.CoinName {
			Balance = ret.Balances[i : i+1]
			break
		}
	}
	ret.Balances = Balance
	if len(ret.Balances) == 0 {
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Balances[0].Amount = decimal.Zero
	for i := 0; i < params.Num; i++ {
		if i >= len(ret.Balances[0].Unspent) {
			break
		}
		for j := i + 1; j < len(ret.Balances[0].Unspent); j++ {
			if ret.Balances[0].Unspent[j].Value.Cmp(ret.Balances[0].Unspent[i].Value) > 0 {
				ret.Balances[0].Unspent[j], ret.Balances[0].Unspent[i] = ret.Balances[0].Unspent[i], ret.Balances[0].Unspent[j]
			}
		}
		ret.Balances[0].Amount = ret.Balances[0].Amount.Add(ret.Balances[0].Unspent[i].Value)
	}
	if len(ret.Balances[0].Unspent) > params.Num {
		ret.Balances[0].Unspent = ret.Balances[0].Unspent[0:params.Num]
	}
	ctx.JSON(http.StatusOK, ret)
	return
}
