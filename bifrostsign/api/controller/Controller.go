package controller

import (
	"bncsign/api/models"
	"bncsign/common/conf"
	"bncsign/common/log"
	"bncsign/common/validator"
	utils "bncsign/utils/bifrost"
	"encoding/json"
	"fmt"
	"github.com/onethefour/common/xutils"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcutil/base58"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/blake2b"
)

type Controller struct {
	Mod models.HdxModel
	Meta *types.Metadata
	GenesisHash string
	BlockNumber uint64
	RuntimeVersion *types.RuntimeVersion
	Cli *utils.Client
}

func (c *Controller) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", c.createAddress)
		group.POST("/sign", c.sign)
		group.POST("/transfer", c.transfer)
		group.POST("/validAddress", c.validAddress)
		group.POST("/getBalance", c.GetBalance)
	}
	go c.update()
}
func (c *Controller) update(){
	var block *types.SignedBlock
	var lastBlockHash types.Hash
	var err error
	c.Cli, err = utils.NewClient(conf.GetConfig().Node.Url)
	//log.Info(conf.GetConfig().Node.Url)
	if err != nil {
		log.Info(err.Error())
		goto end
	}
	c.Meta = c.Cli.Meta

	c.GenesisHash = c.Cli.GetGenesisHash().Hex()
	c.RuntimeVersion, err = c.Cli.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		log.Info(err.Error())
		goto end

	}
	lastBlockHash, err = c.Cli.Api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		log.Info(err.Error())
		goto end
	}
	// params.BlockHash = lastBlockHash.Hex()

	block, err = c.Cli.Api.RPC.Chain.GetBlock(lastBlockHash)
	if err != nil {
		log.Info(err.Error())
		goto end
	}
	c.BlockNumber = uint64(block.Block.Header.Number)
	c.Meta = c.Cli.Meta
end:
	time.Sleep(time.Hour)
	go c.update()
}
func (c *Controller) GetBalance(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		c.NewError(ctx, err.Error())
		return
	}
	for c.Meta != nil {
		time.Sleep(time.Second)
	}

	cli := c.Cli
	acc ,err :=cli.GetAccountInfo(params.Address,c.Meta)
	if err != nil {
		c.NewError(ctx, err.Error())
		return
	}

	ret := new(validator.GetBalanceResponse)


	ret.Data = strconv.FormatInt(acc.Data.Free.Int64(),10)
	ctx.JSON(http.StatusOK, ret)
}

func (c *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		c.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)
	ret.Code = -1
	ret.Data = false

	decodeBytes := base58.Decode(params.Address)
	if len(decodeBytes) != 35 {
		ret.Message = "校验失败,长度不够"
		ctx.JSON(http.StatusOK, ret)
		return
	}

	if decodeBytes[0] != utils.BNCPrefix[0] {
		ret.Message = "prefix valid error"
		ctx.JSON(http.StatusOK, ret)
		return
	}

	pub := decodeBytes[1 : len(decodeBytes)-2]

	data := append(utils.BNCPrefix, pub...)
	input := append(utils.SSPrefix, data...)
	ck := blake2b.Sum512(input)
	checkSum := ck[:2]
	for i := 0; i < 2; i++ {
		if checkSum[i] != decodeBytes[len(decodeBytes)-2+i] {
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
}
func (c *Controller) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
		"data":    "",
	})
}

func (c *Controller) createAddress(ctx *gin.Context) {
	var params = new(validator.CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		c.NewError(ctx, err.Error())
		return
	}

	var returns = &validator.ZcashCreateAddressReturns{
		Data: validator.ZcashCreateAddressReturns_data{CreateAddressParams: *params}}

	addrs, err := c.Mod.NewAccount(params.Num, params.MchName, params.OrderId)
	if err != nil {
		c.NewError(ctx, err.Error())
		return
	}
	returns.Data.Address = addrs
	ctx.JSON(http.StatusOK, returns)
}

func (c *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		c.NewError(ctx, err.Error()+"error")
		return
	}
	if err := xutils.LockMax(params.FromAddress, 3); err != nil {
		log.Info(fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		c.NewError(ctx, fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress, time.Second*5)
	log.Info(String(params))

	for c.Meta == nil {
		time.Sleep(time.Second)
	}
	cli := c.Cli
	meta := c.Meta
	//scanApi := utils.NewDotScanApi(conf.GetConfig().Node.ScanApi,conf.GetConfig().Node.ScanKey)
	//balance,_ ,err := scanApi.AccountInfo(params.FromAddress)
	//if err != nil {
	//	c.NewError(ctx, err.Error())
	//	return
	//}
	acc ,err :=cli.GetAccountInfo(params.FromAddress,meta)
	if err != nil {
		c.NewError(ctx, err.Error())
		return
	}

	log.Info(String(acc))
	return
	//log.Info(balance , params.Amount.IntPart()+273000736)
	if acc.Data.Free.Int64() < params.Amount.IntPart()+273000736 {
		c.NewError(ctx, fmt.Sprintf("地址额度不够:余额%d 出账:%d+273000736",acc.Data.Free.Int64(),params.Amount.IntPart()))
		return
	}
	if params.Nonce == 0 {
		params.Nonce = uint64(acc.Nonce)
	}
	params.GenesisHash = c.GenesisHash
	params.BlockHash = params.GenesisHash
	runVer := c.RuntimeVersion

	params.TransactionVersion = uint32(runVer.TransactionVersion)
	params.SpecVersion = uint32(runVer.SpecVersion)


	params.BlockNumber = c.BlockNumber
	//log.Info(String(params))
	//return

	params.Meta = meta

	var returns = &validator.TelosSignReturns{
		SignHeader: params.SignHeader,
	}

	if rawtx, err := c.Mod.SignTx(params); err != nil {
		c.NewError(ctx, err.Error())
		return
	} else {
		returns.Rawtx = rawtx
		ctx.JSON(http.StatusOK, returns)
		return
	}
}

func (c *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		c.NewError(ctx, err.Error()+"error")
		return
	}
	if err := xutils.LockMax(params.FromAddress, 3); err != nil {
		log.Info(fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		c.NewError(ctx, fmt.Sprintf("from地址:%v交易频繁,未处理,30秒后可重推", params.FromAddress))
		return
	}
	defer xutils.UnlockDelay(params.FromAddress, time.Second*5)
	log.Info(String(params))

	for c.Meta == nil {
		time.Sleep(time.Second)
	}
	cli := c.Cli
	meta := c.Meta
	//scanApi := utils.NewDotScanApi(conf.GetConfig().Node.ScanApi,conf.GetConfig().Node.ScanKey)
	//balance,_ ,err := scanApi.AccountInfo(params.FromAddress)
	//if err != nil {
	//	c.NewError(ctx, err.Error())
	//	return
	//}
	acc ,err :=cli.GetAccountInfo(params.FromAddress,meta)
	if err != nil {
		c.NewError(ctx, err.Error())
		return
	}

	//log.Info(String(acc))
	//return
	//log.Info(balance , params.Amount.IntPart()+273000736)
	if acc.Data.Free.Int64() < params.Amount.IntPart()+273000736 {
		c.NewError(ctx, fmt.Sprintf("地址额度不够:余额%d 出账:%d+273000736",acc.Data.Free.Int64(),params.Amount.IntPart()))
		return
	}
	if params.Nonce == 0 {
		params.Nonce ,err = cli.GetNonce(params.FromAddress)
		if err != nil {
			c.NewError(ctx, err.Error())
			return
		}
	}
	params.GenesisHash = c.GenesisHash
	params.BlockHash = params.GenesisHash
	runVer := c.RuntimeVersion

	params.TransactionVersion = uint32(runVer.TransactionVersion)
	params.SpecVersion = uint32(runVer.SpecVersion)


	params.BlockNumber = c.BlockNumber
	log.Info(String(params))
	//return

	params.Meta = meta

	rawTx, err := c.Mod.SignTx(params)
	if err != nil {
		c.NewError(ctx, err.Error())
		return
	}

	var res interface{}

	log.Info("rawTx:", rawTx)

	err = cli.Api.Client.Call(&res, "author_submitExtrinsic", rawTx)
	if err != nil {
		log.Error(err)
		c.NewError(ctx, err.Error())
		return
	}
	utils.NonceManage.Set(params.FromAddress,int64(params.Nonce+1))
	var returns = &validator.TelosSignReturns{
		SignHeader: params.SignHeader,
	}

	returns.Rawtx = rawTx
	returns.TxHash = res.(string)
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
