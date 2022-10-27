package controller

import (
	"dotsign/api/models"
	"dotsign/common/conf"
	"dotsign/common/log"
	"dotsign/common/validator"
	btc "dotsign/utils/dot"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/blake2b"
	"net/http"
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
		group.POST("/getBalance", this.GetBalance)

	}
	//r.POST("/collector",this.collector)
}
func (this *Controller) GetBalance(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	cli,err := btc.New(conf.GetConfig().Node.Url)
	if err != nil{
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	acc,err := cli.GetAccountInfo(params.Address)
	if err != nil{
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceResponse)
	ret.Data = acc.Data.Free.String()
	ctx.JSON(http.StatusOK, ret)
	return
}
//func (this *Controller) collector(ctx *gin.Context){
//	var params = new(validator.CollectorParams)
//	if err := ctx.ShouldBindJSON(params); err != nil {
//		this.NewError(ctx, err.Error())
//		return
//	}
//	var returns = validator.CollectorResponse{
//		Code: 0,
//		Txs:make([]string,0),
//	}
//	client := mob.NewRpcClient(conf.GetConfig().Node.Url,"","")
//
//	for _,fromaddr := range params.Froms{
//		log.Info(fromaddr)
//		key,err :=this.Mod.GetPrivate(params.MchName,fromaddr)
//		if err != nil {
//			this.NewError(ctx, params.MchName+" addr"+fromaddr+"获取秘钥出错:"+err.Error())
//			return
//		}
//		keys := strings.Split(string(key),"_")
//		monitorid := keys[1]
//		index,_ := strconv.Atoi(keys[3])
//		balance,err :=client.GetBalance(monitorid,int64(index))
//		if err != nil {
//			this.NewError(ctx, params.MchName+" addr"+fromaddr+" getBalance出错:"+err.Error())
//			return
//		}
//		txhash,err := client.SendTransaction(monitorid,balance,params.To,"")
//		returns.Txs = append(returns.Txs,txhash)
//	}
//	ctx.JSON(http.StatusOK, returns)
//	return
//
//}
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
	if len(decodeBytes) != 35 {
		ret.Message = "校验失败,长度不够"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	if decodeBytes[0] != btc.PolkadotPrefix[0] {
		ret.Message = "prefix valid error"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	pub := decodeBytes[1 : len(decodeBytes)-2]

	data := append(btc.PolkadotPrefix, pub...)
	input := append(btc.SSPrefix, data...)
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

	addrs,err :=this.Mod.NewAccount(params.Num,params.MchName,params.OrderId)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	returns.Data.Address=addrs
	ctx.JSON(http.StatusOK, returns)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	client, err := btc.New(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.Meta = client.Meta
	params.GenesisHash = client.GetGenesisHash()
	params.SpecVersion = uint32(client.SpecVersion)
	params.TransactionVersion = uint32(client.TransactionVersion)

	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	if rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawtx
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
	client ,err := btc.New(conf.GetConfig().Node.Url)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	fromAcc,err := client.GetAccountInfo(params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	fromBalance := decimal.NewFromBigInt(fromAcc.Data.Free.Int,0)
	if fromBalance.LessThan(params.Amount){
		this.NewError(ctx, fmt.Sprintf("账户余额不足,出账(%v)<余额(%v)",params.Amount.String(),fromBalance.String()))
		return
	}
	params.Meta = client.Meta
	params.GenesisHash = client.GetGenesisHash()
	params.SpecVersion = uint32(client.SpecVersion)
	params.TransactionVersion= uint32(client.TransactionVersion)
	log.Info(String(params))
	//this.NewError(ctx,"111")
	//return
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	rawtx, err := this.Mod.SignTx(params);
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var result interface{}
	err = client.C.Client.Call(&result, "author_submitExtrinsic", rawtx)
	if err != nil || result == nil {
		this.NewError(ctx, "交易发送失败")
	}
	txid := result.(string)
	returns.Data = rawtx
	returns.TxHash = txid
	ctx.JSON(http.StatusOK, returns)
	return

}

func String(d interface{})string{
	str,_:= json.Marshal(d)
	return string(str)
}