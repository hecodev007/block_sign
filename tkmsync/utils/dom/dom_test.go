package dom

import (
	"testing"
)

var rpc *RpcClient

func TestMain(m *testing.M) {
	rpc = NewRpcClient("https://app-node.domchain.io/")
	m.Run()
}

func TestGetBlockHashByHeight(t *testing.T) {
	h, err := rpc.GetBlockHashByHeight(3)
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(h)
}

func TestGetBlockByHash(t *testing.T) {
	h, err := rpc.GetBlockByHash("0xa2685ab51bfb616db1fb2a0242d760f5fedd5c75561c7990dadcb264a58c5307", true)
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(h.Items[0].Block.Height)
}

func TestGetTransactionByHash(t *testing.T) {
	h, err := rpc.GetTransactionByHash("0xa6ad62fb9cbf579414526e2d4ea2d3b1bd05a46025c98bc986c3074cc73a6c77")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(h.Amount)
}

func TestBlockNumber(t *testing.T) {
	h, err := rpc.BlockNumber()
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(h)
}

func TestGetBlockByHeight(t *testing.T) {
	h, err := rpc.GetBlockByHeight(271967, true)
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(h)
}
