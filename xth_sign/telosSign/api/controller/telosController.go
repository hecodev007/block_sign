package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"telosSign/api/models"
	//"telosSign/utils/eos"
	eos "github.com/eoscanada/eos-go"
)

type TelosController struct {
}

func (this *TelosController) Router(r *gin.Engine) {
	group := r.Group("/v1/tlos")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
	}
}

func (this *TelosController) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
		"data":    "",
	})
}

func (this *TelosController) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &ZcashCreateAddressReturns{
		Data: ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.TelocModel).NewAccount(params.Num, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *TelosController) sign(ctx *gin.Context) {
	var params = new(TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &TelosSignReturns{SignHeader: params.SignHeader}
	if pack, hash, err := new(models.TelocModel).SignTx(params.MchName, params.Data.SignPubKey, params.Data); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		//returns.Data.PackedTransaction
		v, ok := pack.(*eos.PackedTransaction)
		if !ok {
			this.NewError(ctx, "not eos.PackedTransaction")
			return
		}
		returns.Data.PackedTransaction = *v
		returns.Data.TxHash = hash
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *TelosController) transfer(ctx *gin.Context) {
	var params = new(TelosTransferParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &TelosTransferReturns{SignHeader: params.SignHeader}
	eosApi := eos.New("http://telos.rylink.io:20888")
	//eosApi := eos.New("https://mainnet.telosusa.io")
	info, err := eosApi.GetInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.Data.BlockID = info.LastIrreversibleBlockID.String()

	if pack, txhash, err := new(models.TelocModel).SignTx(params.MchName, params.Data.SignPubKey, params.Data); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else if _, err := eosApi.PushTransaction(pack.(*eos.PackedTransaction)); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = txhash
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
