package controller

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"xlmSign/api/models"
	"xlmSign/common/conf"
	"xlmSign/common/log"
	. "xlmSign/common/validator"

	"github.com/onethefour/common/xutils"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"

	"github.com/stellar/go/clients/horizonclient"

	"github.com/gin-gonic/gin"
	hProtocol "github.com/stellar/go/protocols/horizon"
)

type AvaxController struct {
}

func (this *AvaxController) Router(r *gin.Engine) {
	group := r.Group("/v1/" + conf.GetConfig().Name)
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
		group.POST("/transfer", gin.BasicAuth(gin.Accounts{"rylink": "rylink@telos@2020"}), this.transfer)
		group.POST("/trustline", this.trustline)
	}
}

func (this *AvaxController) trustline(ctx *gin.Context) {
	var params = new(TrustLineParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	seed, err := new(models.BiwModel).GetPrivate(params.MchName, params.Address)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	params.Seed = string(seed)
	txhash, err := TrustLine(params)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	ret := new(GetBalanceReturns)
	ret.Code = 0
	ret.Data = txhash
	ret.Message = ""
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
	if returns.Data.Address, err = new(models.BiwModel).NewAccount(params); err != nil {
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
	if Tx, err := new(models.BiwModel).Sign(params); err != nil {
		//fmt.Println(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data, _ = Tx.Base64()
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
	if tx, err := new(models.BiwModel).Sign(params); err != nil {
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	} else {
		returns.Data, err = tx.Base64()
		log.Info("rawtx:")
		log.Info(tx.Base64())
		if err != nil {
			this.NewError(ctx, err.Error())
			return
		}
		client := horizonclient.DefaultPublicNetClient
		var txresp hProtocol.Transaction
		err = errors.New("交易提交失败,联系开发确定是否上链")
		txresp, err := client.SubmitTransactionXDR(returns.Data.(string))
		if err != nil {
			log.Info(err.Error())
			returns.Code = -1
			returns.Message = err.Error()
			txhash, _ := tx.Hash("Public Global Stellar Network ; September 2015")
			returns.TxHash = hex.EncodeToString(txhash[:])
		} else {
			txhash, _ := tx.Hash("Public Global Stellar Network ; September 2015")
			log.Info(hex.EncodeToString(txhash[:]))
			returns.TxHash = txresp.Hash
		}
	}
	log.Info(xutils.String(returns))
	ctx.JSON(http.StatusOK, returns)
	return
}

func TrustLine(params *TrustLineParams) (rawtx string, err error) {
	kp, err := keypair.Parse(params.Seed)
	if err != nil {
		return "", err
	}
	log.Info(kp.Address())
	tokenInfo := strings.Split(params.Token, "-")
	if len(tokenInfo) != 2 {
		return "", errors.New("token错误:" + params.Token)
	}
	client := horizonclient.DefaultPublicNetClient
	sourceAccount, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: kp.Address()})
	code, issuer := strings.ToUpper(tokenInfo[0]), strings.ToUpper(tokenInfo[1])
	log.Info(code, issuer)
	asset := txnbuild.CreditAsset{code, issuer}
	op := txnbuild.ChangeTrust{
		Line:          asset.MustToChangeTrustAsset(),
		SourceAccount: kp.Address(),
		Limit:         "922337203685.4775807",
	}
	transferParams := txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{&op},
		Timebounds:           txnbuild.NewInfiniteTimeout(),
		BaseFee:              txnbuild.MinBaseFee,
		Memo:                 txnbuild.MemoText("test-memo"),
		//EnableMuxedAccounts:  true,
	}
	tx, err := txnbuild.NewTransaction(
		transferParams,
	)
	if err != nil {
		return "", err
	}
	signedTx, err := tx.Sign(network.PublicNetworkPassphrase, kp.(*keypair.Full))
	if err != nil {
		return "", err
	}
	txeBase64, _ := signedTx.Base64()
	//t.Log("Transaction base64: " + txeBase64)

	// Submit the transaction
	resp, err := client.SubmitTransactionXDR(txeBase64)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}
