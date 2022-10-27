package util

import (
	"fmt"
	"testing"
)

func TestRpcClient_SendRequest(t *testing.T) {
	rpc := New("https://dcr.rylink.io:30109", "rylink", "4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=")
	data, err := rpc.SendRequest("validateaddress", []interface{}{"DsmCZKuTHpmN3fMzDpeadx1jJ4iH1zCBar2"})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
