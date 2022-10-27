package stellar

import (
	"stellarsign/common/log"
	"stellarsign/common/validator"
	"strings"

	"github.com/stellar/go/network"

	"github.com/stellar/go/clients/horizonclient"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"

	"errors"

	"github.com/stellar/go/txnbuild"
)

func BuildTx(params *validator.SignParams) (rawtx string, err error) {

	kp, err := seedToKeypire(params.Seed)
	if err != nil {
		return
	}
	var op txnbuild.Operation
	if params.Token == "" {
		op = &txnbuild.Payment{
			Destination:   params.ToAddress,
			Amount:        params.Value.String(),
			Asset:         txnbuild.NativeAsset{},
			SourceAccount: kp.Address(),
		}
	} else {
		tokenInfo := strings.Split(params.Token, "-")
		if len(tokenInfo) != 2 {
			return "", errors.New("token错误:" + params.Token)
		}
		code, issuer := strings.ToUpper(tokenInfo[0]), strings.ToUpper(tokenInfo[1])
		log.Info(code, issuer)
		assert := txnbuild.CreditAsset{code, issuer}
		op = &txnbuild.Payment{
			Destination:   params.ToAddress,
			Amount:        params.Value.String(),
			Asset:         assert,
			SourceAccount: kp.Address(),
		}
	}
	cli := horizonclient.DefaultPublicNetClient
	sourceAccount, err := cli.AccountDetail(horizonclient.AccountRequest{AccountID: kp.Address()})
	if err != nil {
		return "", err
	}
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{op},
			BaseFee:              txnbuild.MinBaseFee,
			Timebounds:           txnbuild.NewInfiniteTimeout(),
			Memo:                 txnbuild.MemoText(params.Memo),
		},
	)

	signedTx, err := tx.Sign(network.PublicNetworkPassphrase, kp)
	if err != nil {
		return "", err
	}
	txeBase64, err := signedTx.Base64()
	if err != nil {
		log.Info(err.Error())
	}
	return txeBase64, err
}
func GetBalance(address string, token string) (amount string, err error) {
	cli := horizonclient.DefaultPublicNetClient
	sourceAccount, err := cli.AccountDetail(horizonclient.AccountRequest{AccountID: address})
	if err != nil {
		return "", err
	}
	tokenInfo := strings.Split(token, "-")
	for _, v := range sourceAccount.Balances {
		if token == "" && v.Type == "native" {
			return v.Balance, nil
		}
		if token != "" && v.Asset.Code == strings.ToUpper(tokenInfo[0]) && v.Asset.Issuer == strings.ToUpper(tokenInfo[1]) {
			return v.Balance, nil
		}
	}
	return "0", nil
}

func TrustLine(params *validator.TrustLineParams) (rawtx string, err error) {
	kp, err := keypair.Parse(params.Seed)
	if err != nil {
		return "", err
	}
	tokenInfo := strings.Split(params.Token, "-")
	if len(tokenInfo) != 2 {
		return "", errors.New("token错误:" + params.Token)
	}
	client := horizonclient.DefaultPublicNetClient
	sourceAccount, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: kp.Address()})
	code, issuer := strings.ToUpper(tokenInfo[0]), strings.ToUpper(tokenInfo[1])
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
func seedToKeypire(seed string) (*keypair.Full, error) {
	seedBytes, err := strkey.Decode(strkey.VersionByteSeed, seed)
	if err != nil {
		return nil, err
	}
	var seed2 [32]byte
	copy(seed2[:], seedBytes)
	key, err := keypair.FromRawSeed(seed2)
	if err != nil {
		return nil, err

	}
	return key, nil
}
