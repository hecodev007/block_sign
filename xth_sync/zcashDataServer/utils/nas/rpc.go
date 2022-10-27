package nas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"zcashDataServer/utils"
)

type NasHttpClient struct {
	client *http.Client
	url    string
}

func NewNasHttpClient(url string) (*NasHttpClient, error) {
	return &NasHttpClient{
		client: http.DefaultClient,
		url:    url,
	}, nil
}

const WEI = 18

const (
	TxNormal   = "binary"
	TxDeploy   = "deploy"
	TxCall     = "call"
	TxProtocol = "protocol"
	TxDip      = "dip"
)

type NodeState struct {
	TailHash     string `json:"tail_hash"`
	LibHash      string `json:"lib_hash"`
	Height       uint64 `json:"height"`
	Synchronized bool   `json:"synchronized"`
	Version      string `json:"version"`
}

type Event struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

type CallData struct {
	Function string `json:"Function"`
	Args     string `json:"Args"`
}
type Erc20Args struct {
	Address string
	Amount  int64
}

type Transaction struct {
	Hash            string   `json:"hash"`
	BlockHash       string   `json:"block_hash"`
	BlockHeight     uint64   `json:"block_height"`
	From            string   `json:"from"`
	To              string   `json:"to"`
	Value           *big.Int `json:"value"`
	Nonce           uint64   `json:"nonce"`
	Timestamp       int64    `json:"timestamp"`
	Type            string   `json:"type"`
	Data            string   `json:"data"`
	GasPrice        *big.Int `json:"gas_price"`
	GasUsed         int64    `json:"gas_used"`
	ContractAddress string   `json:"contract_address"`
	Status          int      `json:"status"`
	ExecuteResult   string   `json:"execute_result"`
	ExecuteError    string   `json:"execute_error"`
}

// Block - block object
type Block struct {
	Hash         string         `json:"hash"`
	ParentHash   string         `json:"parent_hash"`
	Height       uint64         `json:"height"`
	Nonce        uint64         `json:"nonce"`
	Coinbase     string         `json:"coinbase"`
	Timestamp    int64          `json:"timestamp"`
	IsFinality   bool           `json:"is_finality"`
	Transactions []*Transaction `json:"transactions"`
}

type Response struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type ProxyNodeState struct {
	ChainId         uint64 `json:"chain_id,omitempty"`
	Tail            string `json:"tail,omitempty"`
	Lib             string `json:"lib,omitempty"`
	Height          string `json:"height,omitempty"`
	ProtocolVersion string `json:"protocol_version,omitempty"`
	Synchronized    bool   `json:"synchronized,omitempty"`
	Version         string `json:"version,omitempty"`
}

type ProxyTransaction struct {
	Hash            string `json:"hash,omitempty"`
	ChainId         uint32 `json:"chainId,omitempty"`
	From            string `json:"from,omitempty"`
	To              string `json:"to,omitempty"`
	Value           string `json:"value,omitempty"`
	Nonce           string `json:"nonce,omitempty"`
	Timestamp       string `json:"timestamp,omitempty"`
	Type            string `json:"type,omitempty"`
	Data            string `json:"data,omitempty"`
	GasPrice        string `json:"gas_price,omitempty"`
	GasLimit        string `json:"gas_limit,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	Status          int32  `json:"status,omitempty"`
	GasUsed         string `json:"gas_used,omitempty"`
	ExecuteError    string `json:"execute_error,omitempty"`
	ExecuteResult   string `json:"execute_result,omitempty"`
	BlockHeight     string `json:"block_height,omitempty"`
}

func (proxy *ProxyTransaction) toTransaction() *Transaction {
	tx := &Transaction{
		Hash:            proxy.Hash,
		From:            proxy.From,
		To:              proxy.To,
		Type:            proxy.Type,
		Data:            proxy.Data,
		ContractAddress: proxy.ContractAddress,
		Status:          int(proxy.Status),
		ExecuteError:    proxy.ExecuteError,
		ExecuteResult:   proxy.ExecuteResult,
	}
	tx.BlockHeight, _ = utils.ParseUint64(proxy.BlockHeight)
	tx.Nonce, _ = utils.ParseUint64(proxy.Nonce)
	tx.Timestamp, _ = utils.ParseInt64(proxy.Timestamp)
	tx.Value, _ = utils.ParseBigInt(proxy.Value)
	tx.GasPrice, _ = utils.ParseBigInt(proxy.GasPrice)
	tx.GasUsed, _ = utils.ParseInt64(proxy.GasUsed)
	return tx
}

type ProxyBlock struct {
	Hash         string              `json:"hash,omitempty"`
	ParentHash   string              `json:"parent_hash,omitempty"`
	Height       string              `json:"height,omitempty"`
	Nonce        string              `json:"nonce,omitempty"`
	Coinbase     string              `json:"coinbase,omitempty"`
	Timestamp    string              `json:"timestamp,omitempty"`
	ChainId      int                 `json:"chain_id,omitempty"`
	StateRoot    string              `json:"state_root,omitempty"`
	TxsRoot      string              `json:"txs_root,omitempty"`
	EventsRoot   string              `json:"events_root,omitempty"`
	Miner        string              `json:"miner,omitempty"`
	RandomSeed   string              `json:"randomSeed,omitempty"`
	RandomProof  string              `json:"randomProof,omitempty"`
	IsFinality   bool                `json:"is_finality,omitempty"`
	Transactions []*ProxyTransaction `json:"transactions,omitempty"`
}

func (proxy *ProxyBlock) toBlock() *Block {
	block := &Block{
		Hash:       proxy.Hash,
		ParentHash: proxy.ParentHash,
		Coinbase:   proxy.Coinbase,
		IsFinality: proxy.IsFinality,
	}
	block.Height, _ = utils.ParseUint64(proxy.Height)
	block.Nonce, _ = utils.ParseUint64(proxy.Nonce)
	block.Timestamp, _ = utils.ParseInt64(proxy.Timestamp)
	block.Transactions = make([]*Transaction, len(proxy.Transactions))
	for i := range proxy.Transactions {
		block.Transactions[i] = proxy.Transactions[i].toTransaction()
		block.Transactions[i].BlockHash = block.Hash
	}
	return block
}

func (proxy *ProxyNodeState) toNodeState() *NodeState {
	state := &NodeState{
		TailHash:     proxy.Tail,
		LibHash:      proxy.Lib,
		Synchronized: proxy.Synchronized,
		Version:      proxy.Version,
	}

	state.Height, _ = utils.ParseUint64(proxy.Height)
	return state
}

func (c *NasHttpClient) GetNebState() (*NodeState, error) {
	url := c.url + "/v1/user/nebstate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyNodeState{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, err
	}

	return proxy.toNodeState(), nil
}

func (c *NasHttpClient) LatestIrreversibleBlock() (*Block, error) {
	url := c.url + "/v1/user/lib"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyBlock{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, err
	}

	return proxy.toBlock(), nil
}

type GetBlockByHeightRequest struct {
	// block height.
	Height uint64 `json:"height,omitempty"`
	// If true it returns the full transaction objects, if false only the hashes of the transactions.
	FullFillTransaction bool `json:"full_fill_transaction,omitempty"`
}

func (c *NasHttpClient) GetBlockByHeight(height uint64, full bool) (*Block, error) {
	url := c.url + "/v1/user/getBlockByHeight"
	reqdata, err := json.Marshal(&GetBlockByHeightRequest{Height: height, FullFillTransaction: full})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqdata))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyBlock{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, err
	}

	return proxy.toBlock(), nil
}

type GetBlockByHashRequest struct {
	// block height.
	Hash string `json:"hash,omitempty"`
	// If true it returns the full transaction objects, if false only the hashes of the transactions.
	FullFillTransaction bool `json:"full_fill_transaction,omitempty"`
}

func (c *NasHttpClient) GetBlockByHash(hash string, full bool) (*Block, error) {
	url := c.url + "/v1/user/getBlockByHash"
	reqdata, err := json.Marshal(&GetBlockByHashRequest{Hash: hash, FullFillTransaction: full})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqdata))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyBlock{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, err
	}

	return proxy.toBlock(), nil
}

type GetTransactionByHashRequest struct {
	// block height.
	Hash string `json:"hash,omitempty"`
}

func (c *NasHttpClient) GetTransactionReceipt(txid string) (*Transaction, error) {
	url := c.url + "/v1/user/getTransactionReceipt"
	reqdata, err := json.Marshal(&GetTransactionByHashRequest{Hash: txid})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqdata))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyTransaction{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, err
	}

	return proxy.toTransaction(), nil
}

//func (c *NasHttpClient) GetEventsByHash() (*rpcpb.EventsResponse, error) {
//
//	return c.Client.GetEventsByHash(context.Background(), &rpcpb.HashRequest{})
//}

func (c *NasHttpClient) call(req *http.Request) (json.RawMessage, error) {

	//rpc.log.Infof("rpc.client.Do %v \n", req)
	response, err := c.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	//rpc.log.Println(fmt.Sprintf("%s\nResponse: %s\n", method, data))

	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Result, nil
}
func ParseCallData(data []byte) (*CallData, error) {
	got, err := utils.Base64Decode(data)
	if err != nil {
		return nil, err
	}
	var res = &CallData{}
	err = json.Unmarshal(got, res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func ParseTransferData(data []byte) (string, decimal.Decimal, error) {
	var ags = make([]interface{}, 0)
	err := json.Unmarshal(data, &ags)
	if err != nil {
		return "", decimal.Zero, err
	}

	if len(ags) < 2 {
		return "", decimal.Zero, fmt.Errorf("transfer data len less than 2,  %d", len(ags))
	}

	var d decimal.Decimal
	switch ags[1].(type) {
	case string:
		d, err = decimal.NewFromString(ags[1].(string))
		if err != nil {
			return "", decimal.Zero, err
		}
		break
	case float64:
		d = decimal.NewFromFloat(ags[1].(float64))
		break
	default:
		return "", decimal.Zero, fmt.Errorf("don't support type %T", ags[1])
	}

	return ags[0].(string), d, nil
}
