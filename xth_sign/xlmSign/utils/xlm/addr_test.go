package xlm

import (
	"net/http"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
)

func Test_addr(t *testing.T) {
	addr, seed, err := GenAccount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr, seed)
}
func Test_acc(t *testing.T) {
	client := &horizonclient.Client{
		HorizonURL: "http://xlm.rylink.io:31680/",
		HTTP:       http.DefaultClient,
	}
	ar := horizonclient.AccountRequest{AccountID: "GCSC5Y5UE4G3F3AQQCGP4GVIMTQEFA4LX5Q52IBCAT7TDSAOM77NPQWC"}
	sourceAccount, err := client.AccountDetail(ar)
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(sourceAccount)
	t.Log("success")
}

func Test_client(t *testing.T) {
	client := horizonclient.DefaultPublicNetClient
	txBase64 := "AAAAAgAAAAADDYn+VG5b1OaYaiD9RihoZcUnw9G0oKWLcDdQWaLfQwAAAGQB9JBpAAAAMgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAoxODgyODg1NTE1AAAAAAABAAAAAQAAAAADDYn+VG5b1OaYaiD9RihoZcUnw9G0oKWLcDdQWaLfQwAAAAEAAAAAE8DIySepA+QZ7TqKOAF4xYep8txLLTatMUxbmrAawwQAAAAAAAAAAOdpxPgAAAAAAAAAAVmi30MAAABAxCdrKskkKtBV3JKlhWsJdOUm0Rhr3R4H+p48p7uQLKE1QFnSx+Kfn1YaTbkfG3dalRFb6NQfFqlkj6yUPdr2AQ=="
	tx, err := client.SubmitTransactionXDR(txBase64)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(tx.Hash)
}
