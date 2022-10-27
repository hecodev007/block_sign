package cph

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
)

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

type RpcClient struct {
	Client *rpc.Client
}

func NewRpcClient(url, user, pwd string) *RpcClient {
	dial, err := rpc.Dial(url)
	if err != nil {
		panic(err.Error() + " " + url)
	}
	return &RpcClient{
		Client: dial,
	}
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

func (txr *TransactionReceipt) Init() {
	txr.From = ToAddressCypherium(txr.From)
	txr.To = ToAddressCypherium(txr.To)
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
	err := rpc.Client.Call(&result, "cph_getTransactionReceipt", txhash)
	result.Init()
	return &result, err
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

func (tx *Transaction) Init() {
	tx.From = ToAddressCypherium(tx.From)
	tx.To = ToAddressCypherium(tx.To)
}

func (rpc *RpcClient) TransactionByHash(txhash string) (*Transaction, error) {
	var result Transaction
	err := rpc.Client.Call(&result, "cph_getTransactionByHash", txhash)
	result.Init()
	return &result, err
}

func ToCommonAddress(addr string) (address common.Address) {
	addr = strings.Replace(strings.ToLower(addr), "cph", "0x", 1)
	return common.HexToAddress(addr)
}

func (rpc *RpcClient) BlockByNumber(h int64) (*Block, error) {
	var result Block
	err := rpc.Client.Call(&result, "cph_getTxBlockByNumber", hexutil.Uint64(h).String(), true, true)
	for i, _ := range result.Transactions {
		result.Transactions[i].Init()
	}
	return &result, err
}

func (rpc *RpcClient) BlockNumber() (int64, error) {
	var result hexutil.Uint64
	err := rpc.Client.Call(&result, "cph_txBlockNumber")
	return int64(result), err
}
