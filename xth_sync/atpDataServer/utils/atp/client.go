package atp

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"fmt"
)
type Transaction struct{
	Hash string `json:"hash"`
	BlockHash string `json:"blockHash"`
	BlockNumber hexutil.Uint64 `json:"blockNumber"`
	From string `json:"from"`
	To string `json:"to"`
	Value hexutil.Big `json:"value"`
	Gas hexutil.Uint64 `json:"gas"`
	GasPrice hexutil.Uint64 `json:"gasPrice"`
	Input string `json:"input"`
	Nonce hexutil.Uint64  `json:"nonce"`
	TransactionIndex hexutil.Uint64  `json:"transactionIndex"`
	V string `json:"v"`
	R string `json:"r"`
	S string `json:"s"`
}
type TransactionReceipt struct {
	BlockHash string `json:"blockHash"`
	BlockNumber hexutil.Uint64 `json:"blockNumber"`
	ContractAddress string `json:"contractAddress"`
	CumulativeGasUsed hexutil.Uint64 `json:"cumulativeGasUsed"`
	From string `json:"from"`
	To string `json:"string"`
	GasUsed hexutil.Uint64 `json:"gasUsed"`
	Logs []interface{} `json:"logs"`
	Status hexutil.Uint64 `json:"status"`
	TransactionHash string `json:"transactionHash"`
	TransactionIndex hexutil.Uint64 `json:"transactionIndex"`
}
type Block struct {
	Hash string `json:"hash"`
	ParentHash string `json:"parentHash"`
	Number hexutil.Uint64 `json:"number"`
	Timestamp hexutil.Uint64 `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}
type Block2 struct {
	Hash string `json:"hash"`
	ParentHash string `json:"parentHash"`
	Number hexutil.Uint64 `json:"number"`
	Timestamp hexutil.Uint64 `json:"timestamp"`
	Transactions []string `json:"transactions"`
}
func (client *RpcClient)SendRawTransaction(rawTx string) error{
	return client.CallNoAuth("platon_sendRawTransaction",nil,rawTx)
}

func (client *RpcClient)PendingNonceAt(addr string) (uint64,error){
	var result hexutil.Uint64
	err:= client.CallNoAuth("platon_getTransactionCount",&result,addr,"pending")
	return uint64(result),err
}

func (client *RpcClient)SuggestGasPrice() (*big.Int, error) {
	var hex hexutil.Big
	err:= client.CallNoAuth("platon_gasPrice",&hex)
	return (*big.Int)(&hex), err
}
func (client *RpcClient) BalanceAt(addr string) (*big.Int, error) {
	var result hexutil.Big
	err := client.CallNoAuth(  "platon_getBalance", &result,addr, "pending")
	return (*big.Int)(&result), err
}

func (client *RpcClient)TransactionByHash(txhash string)(tx *Transaction, err error){
	var result Transaction
	err = client.CallNoAuth(  "platon_getTransactionByHash", &result,txhash)
	return &result,err
}
func (client *RpcClient)TransactionReceipt(txhash string)(TransactionReceipt, error){
	var result TransactionReceipt
	err := client.CallNoAuth(  "platon_getTransactionReceipt", &result,txhash)
	return result,err
}
func (client *RpcClient)GetBlockByHeight(height int64)(Block, error){
	var result Block
	err := client.CallNoAuth(  "platon_getBlockByNumber", &result, fmt.Sprintf("0x%x",height),true)
	return result,err
}
func (client *RpcClient)BlockByHash(blockHash string)(Block2, error){
	var result Block2
	err := client.CallNoAuth(  "platon_getBlockByHash", &result,blockHash,false)
	return result,err
}

func (client *RpcClient)GetBlockCount()(int64,error){
	var result hexutil.Uint64
	err := client.CallNoAuth(  "platon_blockNumber", &result)
	return int64(result),err
}