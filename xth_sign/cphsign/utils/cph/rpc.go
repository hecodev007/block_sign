package cph

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
)

type Block struct {
	Number *big.Int
}

type RpcClient struct {
	Client *rpc.Client
}

func NewRpcClient(url string) (*RpcClient, error) {
	dial, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		Client: dial,
	}, nil
}

func (rpc *RpcClient) GetBlockCount() (int64, error) {
	var result hexutil.Uint64
	err := rpc.Client.Call(&result, "cph_txBlockNumber")
	return int64(result), err
}

func (rpc *RpcClient) SendRawTransaction(rawtx string) (txhash string, err error) {
	err = rpc.Client.Call(&txhash, "cph_sendRawTransaction", rawtx)
	return txhash, err
}

func (rpc *RpcClient) GetBalance(addr string) (value decimal.Decimal, err error) {
	addr = ToCommonAddress(addr).String()
	result := new(hexutil.Big)
	err = rpc.Client.Call(result, "cph_getBalance", addr, "latest")
	if err != nil {
		return
	}
	bigint := (*big.Int)(result)
	value = decimal.NewFromBigInt(bigint, 0)
	return
}

func (rpc *RpcClient) SuggestGasPrice() (*big.Int, error) {
	result := new(hexutil.Big)
	err := rpc.Client.Call(result, "cph_gasPrice", "latest")
	return (*big.Int)(result), err
}

func (rpc *RpcClient) PendingNonceAt(addr string) (uint64, error) {
	addr = ToCommonAddress(addr).String()
	var result hexutil.Uint64
	err := rpc.Client.Call(&result, "cph_getTransactionCount", addr, "pending")
	return uint64(result), err
}
