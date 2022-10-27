package sol

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/onethefour/common/xutils"
	"github.com/portto/solana-go-sdk/client"
	"github.com/streamingfast/solana-go"
	"github.com/streamingfast/solana-go/programs/token"
	"github.com/streamingfast/solana-go/rpc"
)

func Test_rpc(t *testing.T) {
	erc := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	cli := rpc.NewClient("https://solana-api.projectserum.com")
	//cli.GetProgramAccounts()
	var m token.Mint
	err := cli.GetAccountDataIn(context.Background(), erc, &m)
	if err != nil {
		t.Fatal(err.Error())
	}
	// handle `err`

	json.NewEncoder(os.Stdout).Encode(m)

	// {"OwnerOption":1,
	//  "Owner":"2wmVCSfPxGPjrnMMn7rchp4uaeoTqN39mXFC2zhPdri9",
	//  "Decimals":128,
	//  "IsInitialized":true}
}

func Test_cli(t *testing.T) {
	cli := client.NewClient("https://solana-api.projectserum.com")
	amount, err := cli.GetBalance("FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea")
	if err != nil {
		t.Fatal()
	}
	//cli.GetAccountInfo("FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea")
	t.Log(amount)
	param2 := client.GetAccountInfoConfig{
		Encoding:  "base64",
		DataSlice: client.GetAccountInfoConfigDataSlice{0, 10},
	}
	account, err := cli.GetAccountInfo("FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea", param2)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(account.Owner)
	t.Log(xutils.String(account))
	account, err = cli.GetAccountInfo("8SP6hufbwy7yhJ7jdQPGbeiMLfkvGkLi5CYmYxbRfExm", param2)
	t.Log(account.Owner)
	t.Log(base64.StdEncoding.DecodeString(account.Data.([]interface{})[0].(string)))
	t.Log(xutils.String(account))

}

func Test_client(t *testing.T) {

	cli := NewClient("https://solana-api.projectserum.com")
	//cli.get
	cli.SendRawTransaction(nil)
	t.Log(cli.GetAccountInfo("FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea"))
	t.Log(cli.GetAccountInfo("8SP6hufbwy7yhJ7jdQPGbeiMLfkvGkLi5CYmYxbRfExm"))
}
