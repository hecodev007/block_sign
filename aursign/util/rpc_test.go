package util

import (
	"fmt"
	"testing"
)

func TestRpcClient_SendRequest(t *testing.T) {
	rpc := New("", "", "")
	data, err := rpc.SendRequest("net_version", []interface{}{})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
