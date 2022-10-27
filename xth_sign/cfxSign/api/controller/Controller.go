package controller

import (
	"cfxSign/api/models"
	"cfxSign/common/conf"
	"cfxSign/common/log"
	"cfxSign/common/validator"
	"cfxSign/utils"
	"encoding/hex"
	"math/big"
	"net/http"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Controller struct {
	Mod models.DagModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/validAddress", this.validAddress)
		group.POST("/getBalance", this.getBalance)

	}
}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)

	_, err := cfxaddress.NewFromBase32(params.Address)

	if err != nil {
		ret.Code = -1
		ret.Data = false
		ret.Message = err.Error()
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
	return
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

	var returns = &validator.ZcashCreateAddressReturns{
		Data: validator.ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = this.Mod.NewAccount(params.Num, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	if txhash, rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {

		returns.Data = hex.EncodeToString(rawtx)
		returns.TxHash = txhash
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *Controller) getBalance(ctx *gin.Context) {

	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	addr, _ := cfxaddress.NewFromBase32(params.Address)
	client, err := sdk.NewClient(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturn)
	if params.Contract == "" {

		balance, err := client.GetBalance(addr, nil)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		ret.Data = balance.String()
	} else {
		abijson := "[{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"
		tokenaddress, err := cfxaddress.NewFromBase32(params.Contract)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		erc20, err := sdk.NewContract([]byte(abijson), client, &tokenaddress)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		balance := &struct{ Balance *big.Int }{}
		if err = erc20.Call(nil, balance, "balanceOf", addr.MustGetCommonAddress()); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		ret.Data = balance.Balance.String()
	}
	ctx.JSON(http.StatusOK, ret)
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if !utils.Limit(params.FromAddress, 15) {
		this.NewError(ctx, "limit 1 request per 15s")
		return
	}
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client, err := sdk.NewClient(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	fromAddress, _ := cfxaddress.NewFromBase32(params.FromAddress)
	//toAddress,_:= cfxaddress.NewFromBase32(params.ToAddress)
	if params.Nonce.IsZero() {
		if nonce, err := client.GetNextNonce(fromAddress, nil); err != nil {
			log.Info("error")
			this.NewError(ctx, err.Error())
			return
		} else {
			params.Nonce = decimal.NewFromBigInt((*big.Int)(nonce), 0)
		}
	}

	//
	if params.ChainID == 0 {
		if info, err := client.GetStatus(); err != nil {
			this.NewError(ctx, err.Error())
			return
		} else {
			params.ChainID = uint(info.ChainID)
		}
	}
	//
	if params.EpochHeight == 0 {
		if epoch, err := client.GetEpochNumber(types.EpochLatestState); err != nil {
			this.NewError(ctx, err.Error())
			return
		} else {
			params.EpochHeight = epoch.ToInt().Uint64()
		}
	}
	//gaslimit
	if params.GasLimit.IsZero() {
		if params.Token == "" {
			params.GasLimit = decimal.NewFromInt(21000)
		} else { //代币交易gaslimit
			params.GasLimit = decimal.NewFromInt(100000)
		}
	}
	if params.GasPrice.IsZero() {
		if price, err := client.GetGasPrice(); err != nil {
			this.NewError(ctx, err.Error())
			return
		} else {
			params.GasPrice = decimal.NewFromBigInt((*big.Int)(price), 0)
		}
	}
	if params.Token == "" {
		balance, err := client.GetBalance(fromAddress, nil)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		if params.GasPrice.Mul(params.GasLimit).Add(params.Value).Cmp(decimal.NewFromBigInt((*big.Int)(balance), 0)) > 0 {
			this.NewError(ctx, "insuffient balance,出账额度:"+params.GasPrice.Mul(params.GasLimit).Add(params.Value).Shift(-18).String()+"   实际额度:"+decimal.NewFromBigInt((*big.Int)(balance), -18).String())
			return
		}
	} else {
		abijson := "[{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"
		tokenaddress, err := cfxaddress.NewFromBase32(params.Token)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		erc20, err := sdk.NewContract([]byte(abijson), client, &tokenaddress)
		if err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		balance := &struct{ Balance *big.Int }{}
		if err = erc20.Call(nil, balance, "balanceOf", fromAddress.MustGetCommonAddress()); err != nil {
			log.Info(err.Error())
			this.NewError(ctx, err.Error())
			return
		}
		if params.Value.BigInt().Cmp(balance.Balance) > 0 {
			log.Info("额度不够")
			this.NewError(ctx, "insuffient balance,出账额度:"+params.Value.Shift(-18).String()+"  实际额度:"+decimal.NewFromBigInt(balance.Balance, -18).String())
			return
		}
		//panic("")
	}
	log.Info(params.String())
	if txhash, rawTx, err := this.Mod.SignTx(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else if hash, err := client.SendRawTransaction(rawTx); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Message = "0x" + txhash
		returns.Data = string(hash)
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
