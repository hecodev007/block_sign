package sgb

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	gsrpc "github.com/SubGame-Network/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

func Test_cli(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI("ws://13.231.191.20:9944")
	if err != nil {
		panic(err.Error())
	}
	blockHash, err := types.NewHashFromHexString("0x6020ec58f9b53839c927e760a707323cb6dac8710f1ed94d86ffd9d3c3f22260")
	if err != nil {
		panic(err.Error())
	}
	block, err := api.RPC.Chain.GetBlock(blockHash)
	if err != nil {
		panic(err.Error())
	}
	t.Log(Json(block))
	block, err = api.RPC.Chain.GetBlockLatest()
	t.Log(Json(block))
}

func Test_fee(t *testing.T) {
	t.Log(GetFee("0x13241b3239f7dcbeea8e1b5e18ac944fb6097dc40bcc0624638dc69df084015e"))
}
func Json(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}

func Test_len(t *testing.T) {
	tx, err := hex.DecodeString("3d028400787c794c8a29b900fd0ab5465a83fd5af2a8cc08a76baf7a185d5abfe6ba3d4901a4b63659a21598750581c09357cae4291d3b5ab706f9320442ece3ea2ea1595759a07fbae175b4a748c9668454c413524fdfccfacd296b46df7693e3eaf3e58d0000000a030068608c05b08056223df5695f10d431cec8c38f3d860638a9272741182dc1fd670700e8764817")
	t.Log(len(tx), err)
	//148 0.002820274518
	//147 0.002810312715
	//148 2820818209
	t.Log(2820818209-2732922107, 2732922107+192779000)
	t.Log(900000000000 - 897248857500) //2751140250 2751142500
}
