package algo

import (
	"errors"

	"github.com/algorand/go-algorand-sdk/client/algod"
	almodels "github.com/algorand/go-algorand-sdk/client/algod/models"
)

type Transaction = almodels.Transaction

//包装的RPC-HTTP 客户端
type RpcClient struct {
	algod.Client
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string) *RpcClient {
	rpc := new(RpcClient)
	var err error
	rpc.Client, err = algod.MakeClient(url, "")
	if err != nil {
		panic(err.Error())
	}
	return rpc
}

func (rpc *RpcClient) GetBlockCount() (int64, error) {
	nodestatus, err := rpc.Status()
	if err != nil {
		return 0, err
	}
	if nodestatus.LastRound <= 0 {
		return 0, errors.New("height < 0")
	}
	return int64(nodestatus.LastRound), nil
}

func (rpc *RpcClient) GetBlockByHeight(h int64) (almodels.Block, error) {
	return rpc.Block(uint64(h))
}

func (rpc *RpcClient) TransactionByHash(txid string) (almodels.Transaction, error) {
	return rpc.Client.TransactionByID(txid)
}
