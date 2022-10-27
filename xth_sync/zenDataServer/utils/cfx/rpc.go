package cfx

import (
	"errors"
	"fmt"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"math/big"
	"strings"
)

type Transaction = types.Transaction
type TransactionReceipt = types.TransactionReceipt
type RpcClient struct {
	Client *sdk.Client
}

func NewRpcClient(url, s, k string) *RpcClient {
	client, err := sdk.NewClient(url)
	if err != nil {
		panic(err.Error())
	}
	rpc := &RpcClient{
		Client: client,
	}
	return rpc
}

func (rpc *RpcClient) GetBlockCount() (height int64, err error) {
	status, err := rpc.Client.GetStatus()
	if err != nil {
		return 0, err
	}
	return int64(*status.EpochNumber), nil
}
func (rpc *RpcClient) GetBlockByHeight(h int64) (*types.Block, error) {
	epoch := types.NewEpochNumber(big.NewInt(h))
	return rpc.Client.GetBlockByEpoch(epoch)
}
func (rpc *RpcClient) GetRawTransaction(txhash string) (*types.Transaction, error) {
	return rpc.Client.GetTransactionByHash(types.Hash(txhash))
}
func (rpc *RpcClient) GetBlockByHash(blockHash string) (*types.Block, error) {
	return rpc.Client.GetBlockByHash(types.Hash(blockHash))
}

func (rpc *RpcClient) GetTransactionReceipt(txhash string) (*types.TransactionReceipt, error) {
	return rpc.Client.GetTransactionReceipt(types.Hash(txhash))
}
func (rpc *RpcClient) IsUser(addr string) bool {
	if len(addr) < 4 {
		return false
	}
	return addr[0:3] == "0x1"
	//code, err := rpc.Client.GetCode(types.Address(addr), nil)
	//return len(code) > 2, err

}
func (rpc *RpcClient) IsContract(addr string) bool {
	if len(addr) < 4 {
		return false
	}
	return addr[0:3] == "0x8"
}

func (rpc *RpcClient) IsTransfer(data string) bool {
	if len(data) < 138 {
		return false
	}
	if data[0:10] == "0xa9059cbb" {
		return true
	}
	return false
}
func (rpc *RpcClient) ParseTransferData(input string) (to string, amount *big.Int, err error) {
	//0xa9059cbb0000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000000000000000000000000000000000000005f5e100
	if strings.Index(input, "0xa9059cbb") != 0 {
		return to, amount, errors.New("input is not transfer data")
	}
	if len(input) < 138 {
		return to, amount, fmt.Errorf("input data isn't 138 , size %d ", 138)
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:138], 16)
	if amount.Sign() < 0 {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}
