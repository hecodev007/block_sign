package controller

import (
	"avaxSign/api/models"
	"avaxSign/common/conf"
	"avaxSign/common/log"
	. "avaxSign/common/validator"
	"avaxSign/utils/avax"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/iqoption/zecutil"
	"golang.org/x/crypto/ripemd160"
	"net/http"
)

type AvaxController struct {
}

func (this *AvaxController) Router(r *gin.Engine) {
	group := r.Group("/v1/avax")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.GET("/test", this.test)
	}
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
	if returns.Data.Address, err = new(models.AvaxModel).NewAccount(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
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
	log.Info(string(pjson))
	var returns = &SignReturns{Header: params.Header}
	if rawTx, err := new(models.AvaxModel).Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
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
	if rawTx, err := new(models.AvaxModel).Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawTx
		client := avax.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
		txid, err := client.SendRawTransaction(rawTx)

		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
		} else {
			returns.TxHash = txid
		}
	}

	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *AvaxController) test(ctx *gin.Context) {
	//"address": "GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG",
	//"privkey": "RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"
	//"pubkey": "03ed43dad72976c2b495e1ab9c6ad788e3952e397d602f7bc83453b631a24b224d"
	//	{
	//		"address": "GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG",
	//		"scriptPubKey": "76a914e38837bd9cdd13c1d1ac42fe9a21258c2c6ac58288ac",
	//		"from_ext_address_id": "XV6Ak51EPxndRfJkNueL4ZjFNc2ScQmde7",
	//		"path": "m/0/1",
	//		"ismine": true,
	//		"solvable": true,
	//		"desc": "pkh([e38837bd]03ed43dad72976c2b495e1ab9c6ad788e3952e397d602f7bc83453b631a24b224d)#rxc009nw",
	//		"iswatchonly": false,
	//		"isscript": false,
	//		"iswitness": false,
	//		"pubkey": "03ed43dad72976c2b495e1ab9c6ad788e3952e397d602f7bc83453b631a24b224d",
	//		"iscompressed": true,
	//		"label": "",
	//		"ischange": false,
	//		"labels": [
	//	{
	//		"name": "",
	//		"purpose": "receive"
	//	}
	//]
	//	}

	//	{
	//		"address": "GdVufw2QtMTd2fddoFMMKQq4HyydARsGNX",
	//		"scriptPubKey": "76a914d 7999323a8d240cd816857c17d1bd285c78e3fb 888ac",
	//		"from_ext_address_id": "XV6Ak51EPxndRfJkNueL4ZjFNc2ScQmde7",
	//		"path": "m/0/2",
	//		"ismine": true,
	//		"solvable": true,
	//		"desc": "pkh([d7999323]0363cf6986771e0f4939896a46d282ad5feba2c0c9f347a29d13f76cccaea6dba4)#nrwjrazk",
	//		"iswatchonly": false,
	//		"isscript": false,
	//		"iswitness": false,
	//		"pubkey": "0363cf6986771e0f4939896a46d282ad5feba2c0c9f347a29d13f76cccaea6dba4",
	//		"iscompressed": true,
	//		"label": "",
	//		"ischange": false,
	//		"labels": [
	//	{
	//		"name": "",
	//		"purpose": "receive"
	//	}
	//]
	//	}
	wif, err := btcutil.DecodeWIF("L2yuSWXr4B2x47dALUbYe8fkqEPbgyY8ZL7XWrJpHjtFKwVR8uKw")
	if err != nil {
		//fmt.Println(err.Error())
		return
	}
	pkHash := wif.PrivKey.PubKey().SerializeCompressed()
	//fmt.Printf("%v\n", hex.EncodeToString(pkHash))
	//fmt.Println("hash160", hex.EncodeToString(btcutil.Hash160(pkHash[:ripemd160.Size])))

	var addrPubKey *btcutil.AddressPubKey

	//var decoded = base58.Decode("GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV")
	//fmt.Println("GHOt160", hex.EncodeToString(decoded))
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
	fmt.Println(hex.EncodeToString(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size]))
	address, err := zecutil.EncodeHash(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size], params)
	if err != nil {
		fmt.Println(err.Error())
	}
	if address[0:1] == "G" {
		//fmt.Println(address)
	}
	fmt.Println(address)
	if address == "GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV" {
		fmt.Println(address)
		ctx.JSON(http.StatusOK, address)
		return
	}
	//	}
	//}
	//	}

	ctx.JSON(http.StatusOK, address)

}
