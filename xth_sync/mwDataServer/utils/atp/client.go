package atp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type Transaction struct {
	Hash          string          `json:"transaction"`
	BlockHash     string          `json:"block"`
	BlockNumber   int64           `json:"height"`
	Type          int             `json:"type"`
	From          string          `json:"senderRS"`
	To            string          `json:"recipientRS"`
	Value         decimal.Decimal `json:"amountNQT"`
	Fee           decimal.Decimal `json:"feeNQT"`
	Status        int             `json:"status"`
	Confirmations int64           `json:"confirmations"`
	Deadline      int             `json:"deadline"`
}
type TransactionReceipt struct {
	BlockHash         string         `json:"blockHash"`
	BlockNumber       hexutil.Uint64 `json:"blockNumber"`
	ContractAddress   string         `json:"contractAddress"`
	CumulativeGasUsed hexutil.Uint64 `json:"cumulativeGasUsed"`
	From              string         `json:"from"`
	To                string         `json:"string"`
	GasUsed           hexutil.Uint64 `json:"gasUsed"`
	Logs              []interface{}  `json:"logs"`
	Status            hexutil.Uint64 `json:"status"`
	TransactionHash   string         `json:"transactionHash"`
	TransactionIndex  hexutil.Uint64 `json:"transactionIndex"`
}
type Block struct {
	Hash         string         `json:"block"`
	ParentHash   string         `json:"previousBlock"`
	Height       int64          `json:"height"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
}
type Block2 struct {
	Hash         string         `json:"hash"`
	ParentHash   string         `json:"parentHash"`
	Number       hexutil.Uint64 `json:"number"`
	Timestamp    hexutil.Uint64 `json:"timestamp"`
	Transactions []string       `json:"transactions"`
}

func (client *RpcClient) SendRawTransaction(rawTx string) error {
	return client.CallNoAuth("platon_sendRawTransaction", nil, rawTx)
}

func (client *RpcClient) PendingNonceAt(addr string) (uint64, error) {
	var result hexutil.Uint64
	err := client.CallNoAuth("platon_getTransactionCount", &result, addr, "pending")
	return uint64(result), err
}

func (client *RpcClient) SuggestGasPrice() (*big.Int, error) {
	var hex hexutil.Big
	err := client.CallNoAuth("platon_gasPrice", &hex)
	return (*big.Int)(&hex), err
}
func (client *RpcClient) BalanceAt(addr string) (*big.Int, error) {
	var result hexutil.Big
	err := client.CallNoAuth("platon_getBalance", &result, addr, "pending")
	return (*big.Int)(&result), err
}

func (client *RpcClient) TransactionByHash(txid string) (tx *Transaction, err error) {
	var result Transaction
	url := fmt.Sprintf("%v/%v%v", client.url, "sharder?requestType=getTransaction&transaction=", txid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	return &result, err
}
func (client *RpcClient) TransactionReceipt(txhash string) (TransactionReceipt, error) {
	var result TransactionReceipt
	err := client.CallNoAuth("platon_getTransactionReceipt", &result, txhash)
	return result, err
}
func (client *RpcClient) GetBlockByHeight(height int64) (*Block, error) {
	var result Block
	url := fmt.Sprintf("%v/%v%v%v", client.url, "sharder?requestType=getBlock&height=", height, "&block=&includeTransactions=true")

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	return &result, err

}
func (client *RpcClient) BlockByHash(blockHash string) (Block2, error) {
	var result Block2
	err := client.CallNoAuth("platon_getBlockByHash", &result, blockHash, false)
	return result, err
}

func (client *RpcClient) GetBlockCount() (int64, error) {
	url := fmt.Sprintf("%v/%v", client.url, "sharder?requestType=getBlockchainStatus")
	//println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return gjson.Get(string(body), "lastBlockHeight").Int(), nil
}
