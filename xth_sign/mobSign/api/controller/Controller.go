package controller

import (
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"mobSign/api/models"
	"mobSign/common/conf"
	"mobSign/common/log"
	"mobSign/common/validator"
	"mobSign/utils"
	"mobSign/utils/keystore"
	mob "mobSign/utils/mob"
	"net/http"
	"strconv"
	"strings"
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

	}
	r.POST("/collector",this.collector)
}
func (this *Controller) collector(ctx *gin.Context){
	var params = new(validator.CollectorParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var returns = validator.CollectorResponse{
		Code: 0,
		Txs:make([]string,0),
	}
	client := mob.NewRpcClient(conf.GetConfig().Node.Url,"","")

	for _,fromaddr := range params.Froms{
		log.Info(fromaddr)
		key,err :=this.Mod.GetPrivate(params.MchName,fromaddr)
		if err != nil {
			this.NewError(ctx, params.MchName+" addr"+fromaddr+"获取秘钥出错:"+err.Error())
			return
		}
		keys := strings.Split(string(key),"_")
		monitorid := keys[1]
		index,_ := strconv.Atoi(keys[3])
		balance,err :=client.GetBalance(monitorid,int64(index))
		if err != nil {
			this.NewError(ctx, params.MchName+" addr"+fromaddr+" getBalance出错:"+err.Error())
			return
		}
		txhash,err := client.SendTransaction(monitorid,balance,params.To,"")
		returns.Txs = append(returns.Txs,txhash)
	}
	ctx.JSON(http.StatusOK, returns)
	return

}
func (this *Controller) validAddress(ctx *gin.Context) {
	var params = new(validator.ValidAddressParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.ValidAddressReturns)


	if len(params.Address) != 107 {
		ret.Code = -1
		ret.Data = false
		ret.Message = "长度不对"
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
	var entropy = ""
	var monitorid = ""
	var prelen = int64(0)
	client := mob.NewRpcClient(conf.GetConfig().Node.Url,"","")

	filekey :=	 fmt.Sprintf("%s_c.csv", params.MchName)
	//log.Info(len(keystore.KeysDB[filekey]))
	//this.NewError(ctx, "内部逻辑错误")
	//return
	if len(keystore.KeysDB[filekey]) > 0 {

		for _,v := range keystore.KeysDB[filekey] {
			vs :=strings.Split(v.Key,"_")
			if len(vs) <2 {
				this.NewError(ctx, "内部逻辑错误")
				return
			}
			entropy = vs[0]
			monitorid = vs[1]
			if prelen == 0 {
				if _, prelen, _, _, err = client.GetMonitor(monitorid);err != nil || prelen == 0 {
					this.NewError(ctx, "内部逻辑错误:GetMonitor")
					return
				}
			}

			log.Info(monitorid,prelen)
		}

		if err = client.DelMonitor(monitorid);err != nil{
			this.NewError(ctx, err.Error())
			return
		}
	} else if entropy,err = client.Entropy();err != nil{
		this.NewError(ctx, err.Error())
		return
	}
	if entropy == "" {
		this.NewError(ctx, "内部逻辑错误:entropy生成失败")
		return
	}
	var newlen = prelen+params.Num
	log.Info(prelen,params.Num)
	view_pri ,spend_pri ,err  :=client.GenPri(entropy)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	monitorid,err = client.AddMonitor(view_pri,spend_pri,newlen)
	if err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
	var addrs []string
	for i:=prelen;i< newlen;i++{
		log.Info(i,newlen)
		addr,err :=client.GetAddress(monitorid,i)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		key := fmt.Sprintf("%s_%s_%d_%d",entropy,monitorid,newlen,i)
		addrs = append(addrs, addr)
		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(key), aesKey, true)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: addr, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: addr, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: addr, Key: key})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: addr, Key: ""})
	}
	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	//returns.Data.Address = addrs
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
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if !utils.Limit(params.FromAddress,15){
		this.NewError(ctx, "limit 1 request per 15s")
		return
	}

	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client := mob.NewRpcClient(conf.GetConfig().Node.Url,"","")
	key,err :=this.Mod.GetPrivate(params.MchName,params.FromAddress)
	if err != nil{
		this.NewError(ctx, err.Error())
		return
	}
	keys := strings.Split(string(key),"_")
	if len(keys)<4{
		this.NewError(ctx, "csv key 格式错误")
		return
	}
	monitorid := keys[1]

	txhash,err :=client.SendTransaction(monitorid,params.Value.IntPart(),params.ToAddress,params.Memo)
	if err != nil{
		this.NewError(ctx, err.Error())
		return
	}
	returns.Message = "0x" + txhash
	returns.Data = txhash
	ctx.JSON(http.StatusOK, returns)
	return
}
