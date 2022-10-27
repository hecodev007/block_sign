package wtc

import (
	//"avaxcchainsync/common/log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	//"github.com/onethefour/common/xutils"
	"errors"
)

func (rpc *RpcClient) BlockNumber() (int64, error) {
	var result hexutil.Uint64
	err := rpc.CallNoAuth("eth_blockNumber", &result)
	return int64(result), err
}

type Block struct {
	Number       hexutil.Big    `json:"number"`
	Hash         string         `json:"hash"`
	ParentHash   string         `json:"parentHash"`
	Difficulty   hexutil.Big    `json:"difficulty"`
	GasLimit     hexutil.Big    `json:"gasLimit"`
	GasUsed      hexutil.Big    `json:"gasUsed"`
	Timestamp    hexutil.Uint64 `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct {
	Hash             string         `json:"hash"`
	BlockHash        string         `json:"blockHash"`
	BlockNumber      hexutil.Uint64 `json:"blockNumber"`
	From             string         `json:"from"`
	Gas              hexutil.Uint64 `json:"gas"`
	GasPrice         hexutil.Big    `json:"gasPrice"`
	Input            string         `json:"input"`
	Nonce            hexutil.Uint64 `json:"nonce"`
	To               string         `json:"to"`
	TransactionIndex hexutil.Uint64 `json:"transactionIndex"`
	Value            hexutil.Big    `json:"value"`
}

func (rpc *RpcClient) BlockByNumber(h int64) (*Block, error) {
	var result Block
	err := rpc.CallNoAuth("eth_getBlockByNumber", &result, hexutil.Uint64(h).String(), true)
	return &result, err
}

type TransactionReceipt struct {
	TransactionHash   string         `json:"transactionHash"`
	BlockHash         string         `json:"blockHash"`
	BlockNumber       hexutil.Uint64 `json:"blockNumber"`
	CumulativeGasUsed hexutil.Big    `json:"cumulativeGasUsed"`
	From              string         `json:"from"`
	GasUsed           hexutil.Big    `json:"gasUsed"`
	To                string         `json:"to"`
	Status            hexutil.Uint   `json:"status"`
	Logs              []*Log         `json:"logs"`
}
type Log struct {
	Address     string         `json:"address"`
	Topics      []string       `json:"topics"`
	Data        string         `json:"data"`
	BlockNumber hexutil.Uint64 `json:"blockNumber"`
	LogIndex    hexutil.Uint   `json:"logIndex"`
}

func (rpc *RpcClient) TransactionReceipt(txhash string) (*TransactionReceipt, error) {
	var result TransactionReceipt
	err := rpc.CallNoAuth("eth_getTransactionReceipt", &result, txhash)
	return &result, err
}

func (rpc *RpcClient) TransactionByHash(txhash string) (*Transaction, error) {
	var result Transaction
	err := rpc.CallNoAuth("eth_getTransactionByHash", &result, txhash)
	//result.Value = hexutil.Big(*big.NewInt(1233333333))
	if result.Hash == "" {
		return nil,errors.New("交易不存在")
	}
	//log.Info(xutils.String(result))
	return &result, err
}
