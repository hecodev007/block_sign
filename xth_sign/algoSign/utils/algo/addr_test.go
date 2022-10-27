package algo

import (
	"encoding/json"
	"testing"

	"github.com/algorand/go-algorand-sdk/client/algod"
)

const kmdAddress = "https://algoexplorer.io"
const kmdToken = ""
const algodAddress = "https://api.algoexplorer.io"
const algodToken = "6218386c0d964e371f34bbff4adf543dab14a7d9720c11c6f11970774d4575de"

func Test_addr(t *testing.T) {
	addr, pri, err := GenAccount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)
	t.Log(pri)
}

func Test_client(t *testing.T) {
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		t.Fatal(err.Error())
	}
	params, err := algodClient.SuggestedParams()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(params))
}
func String(v interface{}) string {
	str, _ := json.Marshal(v)
	return string(str)
}
