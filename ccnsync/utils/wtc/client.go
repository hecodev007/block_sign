package wtc

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"solsync/net"
)

func (rpc *RpcClient) BlockNumber() (int64, error) {
	//var result hexutil.Uint64
	//err := rpc.CallNoAuth("eth_blockNumber", &result)
	//return int64(result), err
	var result NetStatus
	param := map[string]string{"action": "status"}
	post, err := net.Post(rpc.url, param)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal([]byte(post), &result)
	if err != nil {
		return 0, err
	}
	if result.Code != 0 || result.Msg != "OK" {
		return 0, errors.New("高度获取失败")
	}
	return result.LastStableBlockIndex, nil
}

func (rpc *RpcClient) BlockByNumber(h int64) (*BlockCCNv2, error) {
	//var result BlockCCN
	//err := rpc.CallNoAuth("eth_getBlockByNumber", &result, hexutil.Uint64(h).String(), true)
	//return &result, err
	param := map[string]interface{}{
		"action": "stable_blocks",
		"limit":  1,
		"index":  h,
	}
	var result BlockCCNv2
	post, err := net.Post(rpc.url, param)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(post), &result)
	if result.Code != 0 || result.Msg != "OK" {
		return nil, errors.New("区块获取失败")
	}
	return &result, nil
}

func (rpc *RpcClient) BlockStatus(hash string) (*BlockStatus, error) {
	//var result BlockCCN
	//err := rpc.CallNoAuth("eth_getBlockByNumber", &result, hexutil.Uint64(h).String(), true)
	//return &result, err
	param := map[string]interface{}{
		"action": "block_state",
		"hash":   hash,
	}
	var result BlockStatus
	post, err := net.Post(rpc.url, param)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(post), &result)
	if result.Code != 0 || result.Msg != "OK" {
		return nil, errors.New("区块状态获取失败")
	}
	return &result, nil
}

type BlockStatus struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	BlockState struct {
		Hash    string `json:"hash"`
		Type    int    `json:"type"`
		Content struct {
			Level int `json:"level"`
		} `json:"content"`
		IsStable      int `json:"is_stable"`
		StableContent struct {
			Status          int64         `json:"status"`
			StableIndex     int64         `json:"stable_index"`
			StableTimestamp int64         `json:"stable_timestamp"`
			Mci             int64         `json:"mci"`
			McTimestamp     int64         `json:"mc_timestamp"`
			FromState       string        `json:"from_state"`
			ToStates        []string      `json:"to_states"`
			GasUsed         string        `json:"gas_used"`
			LogBloom        string        `json:"log_bloom"`
			Log             []interface{} `json:"log"`
			ContractAccount interface{}   `json:"contract_account"`
		} `json:"stable_content"`
	} `json:"block_state"`
}

type NetStatus struct {
	Code                 int64  `json:"code"`
	Msg                  string `json:"msg"`
	Syncing              int64  `json:"syncing"`
	LastStableMci        int64  `json:"last_stable_mci"`
	LastMci              int64  `json:"last_mci"`
	LastStableBlockIndex int64  `json:"last_stable_block_index"`
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

type BlockCCN struct {
	Hash         string         `json:"hash"`
	Parenthash   string         `json:"parentHash"`
	Gaslimit     hexutil.Big    `json:"gasLimit"`
	Gasused      hexutil.Big    `json:"gasUsed"`
	Mingasprice  hexutil.Big    `json:"minGasPrice"`
	Timestamp    hexutil.Uint64 `json:"timestamp"`
	Transactions []string       `json:"transactions"`
	Number       hexutil.Big    `json:"number"`
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

type BlockCCNv2 struct {
	Code      int     `json:"code"`
	Msg       string  `json:"msg"`
	Blocks    []TxCCN `json:"blocks"`
	NextIndex int64   `json:"next_index"`
}

type TxCCN struct {
	Hash    string `json:"hash"`
	Type    int    `json:"type"`
	From    string `json:"from"`
	Content struct {
		To       string          `json:"to"`
		Amount   decimal.Decimal `json:"amount"`
		Previous string          `json:"previous"`
		Gas      decimal.Decimal `json:"gas"`
		GasPrice decimal.Decimal `json:"gas_price"`
		DataHash string          `json:"data_hash"`
		Version  int             `json:"version"`
		Data     string          `json:"data"`
	} `json:"content"`
	Signature string `json:"signature"`
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

func (rpc *RpcClient) TransactionByHash(hash string) (*BlockTx, error) {
	//var result Transaction
	//err := rpc.CallNoAuth("eth_getTransactionByHash", &result, txhash)
	////result.Value = hexutil.Big(*big.NewInt(1233333333))
	//return &result, err

	param := map[string]interface{}{
		"action": "block",
		"hash":   hash,
	}
	var result BlockTx
	post, err := net.Post(rpc.url, param)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(post), &result)
	if result.Code != 0 || result.Msg != "OK" {
		return nil, errors.New("交易获取失败")
	}
	return &result, nil

}

type BlockTx struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Block struct {
		Hash    string `json:"hash"`
		Type    int    `json:"type"`
		From    string `json:"from"`
		Content struct {
			To       string          `json:"to"`
			Amount   decimal.Decimal `json:"amount"`
			Previous string          `json:"previous"`
			Gas      decimal.Decimal `json:"gas"`
			GasPrice decimal.Decimal `json:"gas_price"`
			DataHash string          `json:"data_hash"`
			Version  int             `json:"version"`
			Data     string          `json:"data"`
		} `json:"content"`
		Signature string `json:"signature"`
	} `json:"block"`
}
