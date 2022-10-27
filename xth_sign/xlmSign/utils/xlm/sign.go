package xlm

import (
	"errors"
	"net/http"
	"strings"
	"xlmSign/common/conf"
	"xlmSign/common/log"
	"xlmSign/common/validator"

	"github.com/stellar/go/clients/horizonclient"

	"github.com/stellar/go/network"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
)

func BuildTx(params *validator.SignParams) (tx *txnbuild.Transaction, err error) {

	//sourceAccount := txnbuild.NewSimpleAccount(params.From, int64(1))
	//abcdAsset := txnbuild.CreditAsset{"ABCD", params.From}
	var sourceAccount txnbuild.Account
	client := &horizonclient.Client{
		HorizonURL: conf.Cfg.Node.Url,
		HTTP:       http.DefaultClient,
	}
	if params.Sequence == 0 {
		ar := horizonclient.AccountRequest{AccountID: params.From}
		sourceaccount, err := client.AccountDetail(ar)
		if err != nil {
			log.Info(err.Error())
			return nil, err
		}

		sourceAccount = &sourceaccount
	} else {
		simpleAccount := txnbuild.NewSimpleAccount(params.From, params.Sequence)
		sourceAccount = &simpleAccount
	}
	log.Info("Sequence")
	log.Info(sourceAccount.GetSequenceNumber())
	//sourceAccount1 := txnbuild.NewSimpleAccount(params.From, int64(rand.Uint64()))
	var pathPaymentStrictSend txnbuild.Operation
	if params.Token != "" {
		tokenInfo := strings.Split(params.Token, "-")
		if len(tokenInfo) != 2 {
			return nil, errors.New("token错误:" + params.Token)
		}
		code, issuer := strings.ToUpper(tokenInfo[0]), strings.ToUpper(tokenInfo[1])

		assert := txnbuild.CreditAsset{code, issuer}
		pathPaymentStrictSend = &txnbuild.Payment{
			Destination:   params.To,
			Amount:        params.Value.Shift(-7).String(),
			Asset:         assert,
			SourceAccount: sourceAccount.GetAccountID(),
		}
	} else {
		_, err = client.AccountDetail(horizonclient.AccountRequest{AccountID: params.To})
		if err != nil {
			pathPaymentStrictSend = &txnbuild.CreateAccount{
				Destination:   params.To,
				Amount:        params.Value.Shift(-7).String(),
				SourceAccount: sourceAccount.GetAccountID(),
			}
		} else {
			pathPaymentStrictSend = &txnbuild.Payment{
				Asset:         txnbuild.NativeAsset{},
				Destination:   params.To,
				Amount:        params.Value.Shift(-7).String(),
				SourceAccount: sourceAccount.GetAccountID(),
			}
		}
	}
	fee := params.Fee.IntPart()
	if fee == 0 {
		fee = txnbuild.MinBaseFee
	}

	tx, err = txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{pathPaymentStrictSend},
			BaseFee:              fee,
			Timebounds:           txnbuild.NewInfiniteTimeout(),
			Memo:                 txnbuild.MemoText(params.Memo),
		},
	)
	return tx, err
}

func SignTx(tx *txnbuild.Transaction, seed string) (*txnbuild.Transaction, error) {
	full, err := keypair.ParseFull(seed)
	if err != nil {
		return nil, err
	}
	//log.Info(full.Address())
	tx, err = tx.Sign(network.PublicNetworkPassphrase, full)
	if err != nil {
		return nil, err
	}
	//log.Info(tx.Base64())
	return tx, err
}
