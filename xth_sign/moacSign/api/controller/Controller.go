package controller

import (
	"encoding/hex"
	"encoding/json"
	"moacSign/api/models"
	"moacSign/common/conf"
	"moacSign/common/log"
	"moacSign/common/validator"
	"moacSign/utils"
	btc "moacSign/utils/wtc"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
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
	ret := &validator.ValidAddressReturns{
		Code: -1,
		Data: false,
	}
	if !strings.HasPrefix(params.Address, "0x") {
		ret.Message = "缺少0x前缀"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	if len(params.Address) != 42 {
		ret.Message = "长度不对"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	_, err := hex.DecodeString(strings.TrimPrefix(params.Address, "0x"))
	if err != nil {
		ret.Message = "必须是16进制字符串"
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
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
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	if params.Token == "" {
		balacne, err := client.GetBalance(params.Address)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		ret.Data = balacne.String()
		ctx.JSON(http.StatusOK, ret)
		return
	}
	balacne, err := client.BalanceOf(params.Token, params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret.Data = balacne.String()
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
	addrs, err := this.Mod.NewAccount(int(params.Num), params.MchName, params.OrderId)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	rjson := new(validator.CreateAddressReturns)
	rjson.Data = addrs
	ctx.JSON(http.StatusOK, rjson)
	return
}

func (this *Controller) sign(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	err := ctx.ShouldBindJSON(params)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	if params.Token == "" && params.GasLimit == 0 {
		params.GasLimit = 10000
	} else if params.Token != "" && params.GasLimit == 0 {
		params.GasLimit = 100000
	}

	if params.GasPrice == nil || params.GasPrice.String() == "0" {
		GasPrice, err := client.GasPrice()
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		params.GasPrice = &GasPrice
	}
	if params.GasPrice.Cmp(decimal.NewFromInt(5e10)) > 0 {
		GasPrice := decimal.NewFromInt(5e10)
		params.GasPrice = &GasPrice
	}
	if params.Nonce == 0 {
		params.Nonce, err = client.GetTransactionCount(params.FromAddress, "pending")
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
	}
	log.Info(String(params))
	if txhash, rawtx, err := this.Mod.SignTx(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data = rawtx
		returns.TxHash = txhash
		log.Info(String(returns))
		ctx.JSON(http.StatusOK, returns)
		return
	}
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	err := ctx.ShouldBindJSON(params)
	if err != nil {
		this.NewError(ctx, err.Error()+"error")
		return
	}
	if !utils.Limit(params.FromAddress, 10) {
		this.NewError(ctx, "from地址交易频率限制10秒")
		return
	}
	log.Info(String(params))
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, "", "")
	if params.Token == "" && params.GasLimit == 0 {
		params.GasLimit = 6000
	} else if params.Token != "" && params.GasLimit == 0 {
		params.GasLimit = 15000
	}

	if params.GasPrice == nil || params.GasPrice.String() == "0" {
		GasPrice, err := client.GasPrice()
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		params.GasPrice = &GasPrice
	}
	if params.GasPrice.Cmp(decimal.NewFromInt(50000000000)) > 0 {
		GasPrice := decimal.NewFromInt(50000000000)
		params.GasPrice = &GasPrice
	}
	//检查额度
	if params.Token == "" {
		balance, err := client.GetBalance(params.FromAddress)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		if params.Value.Add(params.GasPrice.Mul(decimal.NewFromInt(int64(params.GasLimit)))).Cmp(balance) > 0 {
			this.NewError(ctx, "账户"+params.FromAddress+"余额不足:"+balance.String()+" 小于 "+params.Value.Add(params.GasPrice.Mul(decimal.NewFromInt(int64(params.GasLimit)))).String())
			return
		}

	} else {
		balance, err := client.BalanceOf(params.Token, params.FromAddress)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		if params.Value.Cmp(balance) > 0 {
			this.NewError(ctx, "代币("+params.TokenName+")余额不足:"+params.Value.String()+"小于"+balance.String())
			//return
		}
	}
	if params.Nonce == 0 {
		params.Nonce, err = client.GetTransactionCount(params.FromAddress, "pending")
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
	}
	log.Info(String(params))
	txhash, rawtx, err := this.Mod.SignTx(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	txid, err := client.SendRawTransaction(rawtx)
	if err != nil {
		log.Info(params.OrderId, txid, txhash, err.Error())
		this.NewError(ctx, err.Error())
		return
	}
	utils.LimitDel(params.FromAddress)
	returns.Rawtx = rawtx
	returns.Txhash = txid
	log.Info(String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
