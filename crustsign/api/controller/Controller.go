package controller

import (
	"crustsign/api/models"
	"crustsign/common/conf"
	"crustsign/common/log"
	"crustsign/common/validator"
	utils "crustsign/utils/crust"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcutil/base58"

	// "github.com/yanyushr/go-substrate-rpc-client/v3/signature"
	"github.com/gin-gonic/gin"

	// "github.com/stafiprotocol/go-substrate-rpc-client/types"

	// sr25519 "github.com/ChainSafe/go-schnorrkel"
	// gsrpc "github.com/yanyushr/go-substrate-rpc-client/v3"
	// "github.com/yanyushr/go-substrate-rpc-client/v3/types"
	// gsrpc "github.com/stafiprotocol/go-substrate-rpc-client"

	"golang.org/x/crypto/blake2b"
)

type Controller struct {
	Mod models.HdxModel
}

func (this *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", this.transfer)
		group.POST("/validAddress", this.validAddress)
		group.POST("/getBalance", this.GetBalance)
	}
}
func (this *Controller) GetBalance(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	cli, err := utils.NewClient(conf.GetConfig().Node.Url)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	acc, err := cli.GetAccountInfo(params.Address, meta)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	ret := new(validator.GetBalanceResponse)

	if acc.Data.Free.Int == nil {
		ret.Data = "0"
	} else {
		ret.Data = acc.Data.Free.String()
	}
	ctx.JSON(http.StatusOK, ret)
	return
}

func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)
	ret.Code = -1
	ret.Data = false

	decodeBytes := base58.Decode(params.Address)
	if len(decodeBytes) != 36 {
		ret.Message = "校验失败,长度不够"
		ctx.JSON(http.StatusOK, ret)
		return
	}

	pre := []byte{(utils.CRustPrefix[0]&0b0000000011111100)>>2 | 0b01000000, utils.CRustPrefix[0]>>8 | (utils.CRustPrefix[0]&0b0000000000000011)<<6}
	if decodeBytes[0] != pre[0] || decodeBytes[1] != pre[1] {
		ret.Message = "prefix valid error"
		ctx.JSON(http.StatusOK, ret)
		return
	}

	pub := decodeBytes[2 : len(decodeBytes)-2]

	data := append(utils.CRustPrefix, pub...)
	input := append(utils.SSPrefix, data...)
	ck := blake2b.Sum512(input)
	checkSum := ck[:2]
	for i := 0; i < 2; i++ {
		if checkSum[i] != decodeBytes[33+i] {
			ret.Message = "checksum valid error"
			ctx.JSON(http.StatusOK, ret)
			return
		}
	}
	if len(pub) != 32 {
		ret.Message = "decode public key length is not equal 32"
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

	addrs, err := this.Mod.NewAccount(params.Num, params.MchName, params.OrderId)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Data.Address = addrs
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	client, err := utils.NewClient(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	meta, err := client.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	params.Meta = meta
	params.GenesisHash = client.GetGenesisHash().Hex()
	params.SpecVersion = uint32(client.SpecVersion)
	params.TransactionVersion = uint32(client.TransactionVersion)

	var returns = &validator.TelosSignReturns{
		SignHeader: params.SignHeader,
	}

	if rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Rawtx = rawtx
		ctx.JSON(http.StatusOK, returns)
		return
	}
}

func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(String(params))

	cli, err := utils.NewClient(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	params.Meta = meta

	params.GenesisHash = cli.GetGenesisHash().Hex()
	params.BlockHash = params.GenesisHash

	runVer, err := cli.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.TransactionVersion = uint32(runVer.TransactionVersion)
	params.SpecVersion = uint32(runVer.SpecVersion)

	lastBlockHash, err := cli.Api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	// params.BlockHash = lastBlockHash.Hex()

	lastBlock, err := cli.Api.RPC.Chain.GetBlock(lastBlockHash)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.BlockNumber = uint64(lastBlock.Block.Header.Number)

	accounrInfo, err := cli.GetAccountInfo(params.FromAddress, meta)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	if !(accounrInfo.Data.Free.Int.Cmp(params.Amount.BigInt()) > 0) {
		this.NewError(ctx, "资金不足")
		return
	}

	params.Nonce = uint64(accounrInfo.Nonce)
	if params.Nonce == 0 {
		params.Nonce = uint64(accounrInfo.Nonce)
	}

	rawTx, err := this.Mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var res interface{}

	fmt.Println("**********rawTx:", rawTx)

	err = cli.Api.Client.Call(&res, "author_submitExtrinsic", rawTx)
	if err != nil {
		log.Error(err)
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &validator.TelosSignReturns{
		SignHeader: params.SignHeader,
	}

	returns.Rawtx = rawTx
	returns.TxHash = res.(string)
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
