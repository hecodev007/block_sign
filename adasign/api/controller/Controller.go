package controller

import (
	"adasign/api/models"
	"adasign/common/conf"
	"adasign/common/log"
	. "adasign/common/validator"
	btc "adasign/utils/ada"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/client"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/onethefour/common/xutils"

	"github.com/gin-gonic/gin"
)

type AvaxController struct {
}

func (this *AvaxController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)

		//group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/transfer", this.transfer)

	}
	r.POST("/api/v1/"+conf.GetConfig().Name+"/unspents", this.unspents)
	r.GET("/api/v1/"+conf.GetConfig().Name+"/validAddress", this.validAddress)
}
func (this *AvaxController) unspents(ctx *gin.Context) {
	var params UnspentsParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(xutils.String(params))
	ret := &UnspentsReturns{}
	cli := btc.NewRpcCli(conf.GetConfig().Node.Url)
	for k, _ := range params {
		coinsRequest := &types.AccountCoinsRequest{
			NetworkIdentifier: btc.NetworkIdentifier,
			AccountIdentifier: &types.AccountIdentifier{
				Address: params[k],
			},
			IncludeMempool: true,
		}
		coins, cerr, err := cli.AccountAPI.AccountCoins(context.Background(), coinsRequest)
		if err != nil {
			ret.Code = 1
			ret.Message = err.Error()
			ctx.JSON(http.StatusOK, ret)
			return
		} else if cerr != nil {
			ret.Code = 1
			ret.Message = cerr.Message
			ctx.JSON(http.StatusOK, ret)
			return
		}

		for ck, _ := range coins.Coins {
			utxo := &Utxo{Tokens: make(map[string]uint64)}

			txinfos := strings.Split(coins.Coins[ck].CoinIdentifier.Identifier, ":")
			if len(txinfos) != 2 {
				continue
			}
			utxo.Txid = txinfos[0]
			utxo.Vout, err = strconv.Atoi(txinfos[1])
			if err != nil {
				log.Info(params[ck] + err.Error())
				continue
			}

			//if
			tmpvalue, _ := strconv.Atoi(coins.Coins[ck].Amount.Value)
			utxo.Tokens[coins.Coins[ck].Amount.Currency.Symbol] = uint64(tmpvalue)
			utxo.Address = params[k]
			//decimals, err := strconv.Atoi(coins[ck].Amount.Currency.Decimals)
			if len(coins.Coins[ck].Amount.Metadata) > 1 {
				this.NewError(ctx, "unspents meta 解析错误:address "+params[ck])
				return
			}
			//log.Info(xutils.String())
			if len(coins.Coins[ck].Amount.Metadata) == 1 {
				tokenMetas, ok := coins.Coins[ck].Amount.Metadata[coins.Coins[ck].CoinIdentifier.Identifier]
				if !ok {
					this.NewError(ctx, "unspents meta.txid-index 解析错误:address "+params[ck])
					return
				}
				tokenMetaBytes, _ := json.Marshal(tokenMetas)
				var ToKenInsts = make([]*btc.TokenBundle, 0)
				if err = json.Unmarshal(tokenMetaBytes, &ToKenInsts); err != nil {
					this.NewError(ctx, err.Error())
					return
				}

				for _, tmpToKenInst := range ToKenInsts {
					if len(tmpToKenInst.Tokens) != 1 {
						this.NewError(ctx, "unspents meta.txid-index.tokens 解析错误:address "+params[ck])
						return
					}
					tokenassertid := fmt.Sprintf("%v-%v-%v", tmpToKenInst.PolicyId, tmpToKenInst.Tokens[0].Currency.Symbol, tmpToKenInst.Tokens[0].Currency.Decimals)
					tmpvalue, _ = strconv.Atoi(tmpToKenInst.Tokens[0].Value)
					utxo.Tokens[tokenassertid] = uint64(tmpvalue)
				}

			}
			ret.Data = append(ret.Data, utxo)
		}
	}
	ctx.JSON(http.StatusOK, ret)
}

func (this *AvaxController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
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
	if returns.Data.Address, err = new(models.DagModel).NewAccount(params.Num, params.MchId, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *AvaxController) validAddress(ctx *gin.Context) {
	var params = new(ValidAddressParams)
	params.Address = ctx.Query("address")
	returns := new(ValidAddressReturns)
	//if len(params.Address) != 100 && len(params.Address) != 58 && len(params.Address) != 103 && len(params.Address) != 104 {
	//	returns.Data.Isvalid = false
	//	ctx.JSON(http.StatusOK, returns)
	//	return
	//}
	//
	clientCfg := client.NewConfiguration(
		conf.GetConfig().Node.Url,
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	var totalInput = make(map[string]uint64)
	op, _ := btc.NewOutputsOperation(params.Address, totalInput)
	preprocessRequest := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: btc.NetworkIdentifier,
		Operations:        []*types.Operation{op},
	}
	_, _, err := cli.ConstructionAPI.ConstructionPreprocess(ctx, preprocessRequest)
	if err != nil {
		log.Info(err.Error())
		returns.Code = 1
		returns.Data.Isvalid = false
		returns.Message = "校验地址失败"
		ctx.JSON(http.StatusOK, returns)
		return
	}
	returns.Data.Isvalid = true
	returns.Code = 0
	returns.Message = ""
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *AvaxController) sign(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	//log.Infof("Ada 发送结构：%+v", params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if txhash, rawTx, err := new(models.DagModel).Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		returns.TxHash = txhash
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *AvaxController) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	pjson, _ := json.Marshal(params)
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if txid, rawTx, err := new(models.DagModel).Sign(params); err != nil {
		log.Info(rawTx)
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		log.Info(txid, rawTx)
		returns.Data = rawTx
		client := btc.NewRpcCli(conf.GetConfig().Node.Url)
		submitRequest := &types.ConstructionSubmitRequest{
			NetworkIdentifier: btc.NetworkIdentifier,
			SignedTransaction: rawTx,
		}
		submitResp, cerr, err := client.ConstructionAPI.ConstructionSubmit(context.Background(), submitRequest)
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else if cerr != nil {
			returns.Code = -1
			returns.Message = cerr.Message
		} else {
			returns.TxHash = submitResp.TransactionIdentifier.Hash
		}
	}
	ctx.JSON(http.StatusOK, returns)
	return
}
