package controller

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"net/http"
	"fioprotocol/api/models"
	"github.com/onethefour/common/xutils"
	"strings"
	"time"
	"fmt"
	//"fioprotocol/api/models"
	"fioprotocol/common/conf"
	"fioprotocol/common/log"
	"fioprotocol/common/validator"

	//tokentypes "github.com/okex/exchain-go-sdk/module/token/types"
	//tokentypes "github.com/okex/exchain/x/token/types"
	extypes "github.com/okex/exchain/x/evm/types"
	//sdk "github.com/okex/exchain-go-sdk/types"
	//sdk "github.com/okex/exchain-go-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//tokentypes "github.com/okex/exchain/x/token/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	btc "fioprotocol/utils/bos"
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
		log.Info(fmt.Sprintf("from??????:%v????????????,?????????,30???????????????", params.FromAddress))
		this.NewError(ctx, fmt.Sprintf("from??????:%v????????????,?????????,30???????????????", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress, time.Second*3)

	if params.ToAddress == "" {
		params.Gaslimit=21000
	} else {
		params.Gaslimit = 100000
	}


	client := btc.NewRpcClient(conf.GetConfig().Node.Eth,"","")
	if params.Nonce == 0 {
		fromAddr, err := sdk.AccAddressFromBech32(params.FromAddress)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}

		Nonce,err := client.GetTransactionCount(common.BytesToAddress(fromAddr.Bytes()).String(),"pending")
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		params.Nonce=Nonce
	}
	if params.Gasprice.String() == "0" {
		price,err := client.GasPrice()
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		params.Gasprice=price.Mul(decimal.NewFromFloat(1.1))
	}
	log.Info(String(params))
	if params.Gasprice.Cmp(decimal.NewFromFloat(100000000000))>0{
		this.NewError(ctx, "tx.fee > 0.1 okt price:"+params.Gasprice.String())
		return
	}
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	prikey,err := ethcrypto.HexToECDSA(string(pri))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	toAddr, err := sdk.AccAddressFromBech32(params.ToAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	//msg := tokentypes.NewMsgTokenSend(fromAddr, toAddr, coins)
	toaddress := common.BytesToAddress(toAddr.Bytes())
	tx := extypes.NewMsgEthereumTx(params.Nonce,&toaddress,params.Value.Shift(18).BigInt(),params.Gaslimit,params.Gasprice.BigInt(),nil)
	err = tx.Sign(big.NewInt(66),prikey)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var tmpbuf []byte
	buff := bytes.NewBuffer(tmpbuf)

	err = tx.EncodeRLP(buff)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Rawtx = "0x" +hex.EncodeToString(buff.Bytes())
	returns.Data =  tx.String()

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
	//???????????????????????????
	log.Info(xutils.String(params))
	//??????????????????,???from????????????,??????nonce??????
	//max???????????????????????????????????????????????????????????????,??????????????????
	if err := xutils.LockMax(params.FromAddress,3);err != nil {
		this.NewError(ctx, params.OrderId+" "+fmt.Sprintf("from??????:%v????????????,?????????,30???????????????", params.FromAddress))
		return
	}
	//??????3?????????,???????????????????????????????????????,nonce??????????????????,???????????????
	defer xutils.UnlockDelay(params.FromAddress,time.Second*3)

	//????????????
	if params.ToAddress == "" {
		params.Gaslimit=21000
	} else {
		params.Gaslimit = 100000
	}

	fromAddress := ""
	fromAddr, err := sdk.AccAddressFromBech32(params.FromAddress)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}
	fromAddress = common.BytesToAddress(fromAddr.Bytes()).String()
	client := btc.NewRpcClient(conf.GetConfig().Node.Eth,"","")
	if params.Nonce == 0 {
		Nonce,err := client.GetTransactionCount(fromAddress,"pending")
		if err != nil {
			this.NewError(ctx, params.OrderId+" "+err.Error())
			return
		}
		params.Nonce=Nonce
	}
	if params.Gasprice.String() == "0" {
		price,err := client.GasPrice()
		if err != nil {
			this.NewError(ctx, params.OrderId+" "+err.Error())
			return
		}
		params.Gasprice=price.Mul(decimal.NewFromFloat(1.1))
	}

	//?????????????????????
	if params.Gasprice.Mul(decimal.NewFromInt(int64(params.Gaslimit))).Cmp(decimal.NewFromInt(1e17)) > 0 {
		this.NewError(ctx, params.OrderId+" "+"?????????????????????,0.1okt")
		return
	}
	//????????????
	if params.ContractAddress == "" {
		balance, err := client.GetBalance(fromAddress,"")
		if err != nil {
			this.NewError(ctx, params.OrderId+" "+err.Error())
			return
		}
		if params.Value.Add(params.Gasprice.Mul(decimal.NewFromInt(int64(params.Gaslimit)))).Cmp(balance) > 0 {
			this.NewError(ctx, params.OrderId+" "+"??????"+params.FromAddress+"????????????:"+balance.String()+" ?????? "+params.Value.Add(params.Gasprice.Mul(decimal.NewFromInt(int64(params.Gaslimit)))).String())
			return
		}

	} else {
		balance, err := client.BalanceOf(params.ContractAddress, fromAddress)
		if err != nil {
			this.NewError(ctx, params.OrderId+" "+err.Error())
			return
		}
		if params.Value.Cmp(balance) > 0 {
			this.NewError(ctx, params.OrderId+" "+"??????("+params.Token+")????????????:"+params.Value.String()+"??????"+balance.String())
			return
		}
	}
	//?????????????????????????????????
	//params.OrderId ?????????request,???????????????????????????
	log.Info(params.OrderId,String(params))
	if params.Gasprice.Cmp(decimal.NewFromFloat(100000000000))>0{
		this.NewError(ctx, params.OrderId+" "+"tx.fee > 0.1 okt price:"+params.Gasprice.String())
		return
	}
	var returns = &validator.SignReturns{SignHeader: params.SignHeader}
	pri, err := this.mod.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}
	prikey,err := ethcrypto.HexToECDSA(string(pri))
	if err != nil {
		this.NewError(ctx,params.OrderId+" "+err.Error())
		return
	}

	toAddr, err := sdk.AccAddressFromBech32(params.ToAddress)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}

	toaddress := common.BytesToAddress(toAddr.Bytes())
	var amount *big.Int
	var paload []byte

	if params.ContractAddress != "" {
		recipient := toaddress
		toaddress = common.HexToAddress(params.ContractAddress)
		amount = big.NewInt(0)
		datastr := "a9059cbb000000000000000000000000" + strings.TrimPrefix(recipient.Hex(), "0x")
		valueByte := params.Value.BigInt().Bytes()
		valuehex := hex.EncodeToString(valueByte)
		valueparam := "0000000000000000000000000000000000000000000000000000000000000000"
		valueparam = valueparam[0:64-len(valuehex)] + valuehex
		datastr += valueparam
		if len(datastr) != 136 {
			this.NewError(ctx,params.OrderId+" "+"????????????data????????????")
			return
		}
		paload, _ = hex.DecodeString(datastr)
	} else {
		amount = params.Value.BigInt()
	}

	tx := extypes.NewMsgEthereumTx(params.Nonce,&toaddress,amount,params.Gaslimit,params.Gasprice.BigInt(),paload)
	err = tx.Sign(big.NewInt(66),prikey)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}
	var tmpbuf []byte
	buff := bytes.NewBuffer(tmpbuf)

	err = tx.EncodeRLP(buff)
	if err != nil {
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}
	returns.Rawtx = "0x" +hex.EncodeToString(buff.Bytes())
	returns.Data =  tx.String()
	txid,err := client.SendRawTransaction(returns.Rawtx)
	if err != nil {
		log.Info(xutils.String(returns))
		this.NewError(ctx, params.OrderId+" "+err.Error())
		return
	}

	returns.Data = txid
	//????????????????????????,?????????rawtx,???????????????????????????????????????
	log.Info(xutils.String(returns))
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
	ret.Code=0
	ret.Data = "0"
	ret.Message = ""

	var balance decimal.Decimal
	var err error
	if params.Token == "" {
		client := btc.NewRpcClient(conf.GetConfig().Node.Eth2,"","")
		balance,err =client.GetBalance(params.Address,"pending")
	} else {
		client := btc.NewRpcClient(conf.GetConfig().Node.Eth,"","")
		balance ,err =client.BalanceOf(params.Token,params.Address)
	}
	if err != nil {
		ret.Code=1
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	} else {
		ret.Data = balance.String()
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
	if len(params.Address) != 41 || !strings.HasPrefix(params.Address, "ex") {
		ret.Code = -1
		ret.Data = false
		ret.Message = "failed. invalid address"
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
