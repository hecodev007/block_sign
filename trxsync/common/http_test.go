package common

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestHttpPost(t *testing.T) {
	req := HttpPost("https://api.nileex.io/wallet/getnowblock")
	blockBytes, err := req.Bytes()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(blockBytes))
}

func TestWebsocket(t *testing.T) {
	rpc, _ := NewRpcClient("ws://127.0.0.1:11212", "", "")

	var resp interface{}
	params := make(map[string]interface{})
	params["length"] = 12
	err := rpc.Post("create_mnemonics", &resp, params)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := json.Marshal(resp)
	fmt.Println(string(d))
}
