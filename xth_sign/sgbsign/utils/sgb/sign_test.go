package sgb

import (
	gcli "github.com/SubGame-Network/go-substrate-rpc-client"
	"testing"
)
var nodeurl = "ws://13.231.191.20:9944"
func Test_sign(t *testing.T){
	api,err := gcli.NewSubstrateAPI(nodeurl)
	if err != nil {
		t.Fatal(err.Error())
	}
	//api.RPC.
}