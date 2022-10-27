package stellar

import (
	"testing"

	"github.com/stellar/go/network"

	"github.com/stellar/go/txnbuild"

	"github.com/onethefour/common/xutils"

	"github.com/stellar/go/keypair"

	"github.com/stellar/go/clients/horizonclient"
)

func Test_cli(t *testing.T) {
	kp, err := keypair.Parse("SBRTFQYUTWONOQRLFSXZFCUFG3XH7D4NDHYMCGR2J7Q3H4XG6FQQHAAM")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(kp.Address())
	client := horizonclient.DefaultPublicNetClient
	ar := horizonclient.AccountRequest{AccountID: kp.Address()}
	sourceAccount, err := client.AccountDetail(ar)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(sourceAccount.AccountID, sourceAccount.Sequence)
	t.Log(xutils.String(sourceAccount.Balances))
	return
	asset := txnbuild.CreditAsset{"LSP", "GAB7STHVD5BDH3EEYXPI3OM7PCS4V443PYB5FNT6CFGJVPDLMKDM24WK"}
	//asset := txnbuild.NativeAsset{}
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
		t.Fatal(err.Error())
	}
	signedTx, err := tx.Sign(network.PublicNetworkPassphrase, kp.(*keypair.Full))
	if err != nil {
		t.Fatal(err.Error())
	}
	txeBase64, _ := signedTx.Base64()
	t.Log("Transaction base64: " + txeBase64)

	// Submit the transaction
	resp, err := client.SubmitTransactionXDR(txeBase64)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(resp))
}
