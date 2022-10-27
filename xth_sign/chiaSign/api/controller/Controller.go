package controller

import (
	"chiaSign/api/models"
	"chiaSign/common/conf"
	"chiaSign/common/log"
	"chiaSign/common/validator"
	"chiaSign/utils"
	btc "chiaSign/utils/chia"
	"chiaSign/utils/keystore"
	"encoding/json"
	"fmt"
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
func (this *Controller) getBalance(ctx *gin.Context) {
	var params = new(validator.GetBalanceParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}

	client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.Cert, conf.GetConfig().Node.Key)
	walletInfo, err := client.WalletInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	walletBalance, err := client.WalletBalance(walletInfo.Wallets[0].Id)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(validator.GetBalanceReturns)

	ret.Code = 0
	ret.Data = decimal.NewFromInt(walletBalance.WalletBalance.Confirmed_wallet_balance).String()
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

	//xch14xlhx28xad689729jhgd5yxy5z3q7tjy4n2gz0xjnkga6ce350kska7pnm
	if !strings.HasPrefix(params.Address, "xch") || len(params.Address) != 62 {
		ret.Code = -1
		ret.Data = false
		ret.Message = "chia地址校验失败:" + params.Address
		ctx.JSON(http.StatusOK, ret)
		return
	}
	ret.Code = 0
	ret.Data = true
	ctx.JSON(http.StatusOK, ret)
	return
}
func (this *Controller) NewError(ctx *gin.Context, err string) {
	log.Info(err)
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
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.Cert, conf.GetConfig().Node.Key)

	var returns = &validator.ZcashCreateAddressReturns{
		Data: validator.ZcashCreateAddressReturns_data{CreateAddressParams: *params}}
	var monic string
	if len(keystore.KeysDB) == 0 {
		monicResponse, err := client.GenerateMnemonic()
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		monic = strings.Join(monicResponse.Mnemonic, " ")
		_, err = client.AddMonic(monic)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
	} else {
		for filekey, csvdb := range keystore.KeysDB {
			if !strings.HasSuffix(filekey, "_c.csv") {
				continue
			}
			for _, v := range csvdb {
				monic = v.Key
				break
			}
			break
		}
	}
	walletInfo, err := client.WalletInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
	var addrs []string
	for i := 0; i < params.Num; i++ {
		addr, err := client.Get_next_address(walletInfo.Wallets[0].Id)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		addrs = append(addrs, addr)
		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(monic), aesKey, true)
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: addr, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: addr, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: addr, Key: monic})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: addr, Key: ""})
	}
	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderId); err != nil {
		this.NewError(ctx, fmt.Sprintf("generateCvsABC err: %v", err))
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
	var returns = &validator.TelosSignReturns{SignHeader: params.SignHeader}

	ctx.JSON(http.StatusOK, returns)
	return
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(validator.TelosSignParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	if !utils.Limit(params.FromAddress, 15) {
		this.NewError(ctx, "limit 1 request per 15s")
		return
	}
	var returns = &validator.TelosTransferReturns{SignHeader: params.SignHeader}
	client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.Cert, conf.GetConfig().Node.Key)
	walletInfo, err := client.WalletInfo()
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	_, err = new(models.DagModel).GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	send, err := client.SendTransaction(walletInfo.Wallets[0].Id, params.Value.IntPart(), params.ToAddress, params.Fee.IntPart())
	if err != nil {
		this.NewError(ctx, "交易发送错误,需要人工确认:"+err.Error())
		return
	}
	if !send.Success {
		this.NewError(ctx, "交易发送失败,需要人工确认")
		return
	}
	log.Info(params.OrderId, String(send))
	returns.Data = send.Transaction_id
	log.Info(params.OrderId, String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
