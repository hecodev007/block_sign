package controller

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ghostSign/api/models"
	"ghostSign/common/conf"
	"ghostSign/common/log"
	"ghostSign/utils/btc"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gin-gonic/gin"
	"github.com/iqoption/zecutil"
	"golang.org/x/crypto/ripemd160"
	"net/http"
)

type GhostController struct {
}

func (this *GhostController) Router(r *gin.Engine) {
	group := r.Group("/v1/ghost")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		//group.GET("/test", this.test)
	}
}

func (this *GhostController) NewError(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": errMsg,
	})
}
func (this *GhostController) createAddress(ctx *gin.Context) {
	var params = new(CreateAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	var returns = &CreateAddressReturns{
		Data: CreateAddressReturns_data{CreateAddressParams: *params}}

	var err error
	if returns.Data.Address, err = new(models.GhostModel).NewAccount(params.Num, params.MchName, params.OrderNo); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) sign(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.GhostModel).HotSignTx(params.MchName, params.SignParams_data); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
	}

	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *GhostController) transfer(ctx *gin.Context) {
	var params = new(SignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	log.Info(Json(params))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.GhostModel).HotSignTx(params.MchName, params.SignParams_data); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		txid, err := client.SendRawTransaction(rawTx)
		if err != nil {
			returns.Code = -1
			log.Info(err.Error())
			returns.Message = err.Error()
		} else {
			returns.TxHash = txid
		}
	}
	log.Info(Json(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *GhostController) test(ctx *gin.Context) {
	addr1 :=pritoaddr("L2yuSWXr4B2x47dALUbYe8fkqEPbgyY8ZL7XWrJpHjtFKwVR8uKw")
	log.Info(addr1)//GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV
	addr1 = pritoaddr("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx")
	log.Info(addr1)//GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG
	addr1 = pritoaddr("RefopvCPh8uH3pWJDKYVb9opG8zNTTmJ2xB3XPKYETH23Wb1dRSw")
	log.Info(addr1)//GbuqcUeaYM4FQTBm3KzhKbpWUR83fSAfY1

	wif, err := btcutil.DecodeWIF("L2yuSWXr4B2x47dALUbYe8fkqEPbgyY8ZL7XWrJpHjtFKwVR8uKw")
	if err != nil {
		log.Info(err.Error())
		return
	}
	pkHash := wif.PrivKey.PubKey().SerializeCompressed()
	log.Infof("%v\n", hex.EncodeToString(pkHash))
	log.Info("hash160", hex.EncodeToString(btcutil.Hash160(pkHash[:ripemd160.Size])))

	var addrPubKey *btcutil.AddressPubKey

	var decoded = base58.Decode("GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV")
	log.Info("GHOt160", hex.EncodeToString(decoded))
	//for st:=0;st<256;st++ {
	//	fmt.Println("st:",st)
	//	chaincfg.MainNetParams.PubKeyHashAddrID = byte(st)
	if addrPubKey, err = btcutil.NewAddressPubKey(pkHash, &chaincfg.MainNetParams); err != nil {
		fmt.Println(err.Error())
		return
	}

	params := []byte{0x1C}
	//for i := 0; i < 256; i++ {
	//	for j := 0; j < 256; j++ {
	params[0] = byte(int8(38))
	log.Info(hex.EncodeToString(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size]))
	address, err := zecutil.EncodeHash(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size], params)
	if err != nil {
		log.Info(err.Error())
	}
	if address[0:1] == "G" {
		//fmt.Println(address)
	}
	log.Info(address)
	if address == "GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV" {
		log.Info(address)
		ctx.JSON(http.StatusOK, address)
		return
	}
	//	}
	//}
	//	}

	ctx.JSON(http.StatusOK, address)

}
func pritoaddr(wifstr string) string{
	params := []byte{0x1C}
	params[0] = byte(int8(38))
	wif, err := btcutil.DecodeWIF(wifstr)
	if err !=nil {
		panic(err.Error())
	}
	pkHash := wif.PrivKey.PubKey().SerializeCompressed()
	//pkHash := wif.PrivKey.PubKey().SerializeUncompressed()
	 addrPubKey, err := btcutil.NewAddressPubKey(pkHash, &chaincfg.MainNetParams)
	if err !=nil {
		panic(err.Error())
	}
	address, err := zecutil.EncodeHash(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size], params)
	if err !=nil {
		panic(err.Error())
	}
	return address
}

func Json(d interface{}) string{
	str,_:= json.Marshal(d)
	return string(str)
}