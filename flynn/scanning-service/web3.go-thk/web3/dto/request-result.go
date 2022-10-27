package dto

import (
	"errors"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/complex/types"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/constants"
	"strconv"
	"strings"

	"encoding/json"
	"fmt"
	"math/big"
)

type RequestResult struct {
	// ID      int         `json:"id"`
	// Version string      `json:"jsonrpc"`
	Result interface{} `json:"result"`
	Error  *Error      `json:"error,omitempty"`
	Data   string      `json:"data,omitempty"`
}

type BalanceResult struct {
	Address     string        `json:"address"`
	Nonce       int           `json:"nonce"`
	Balance     types.Uint128 `json:"balance"`
	StorageRoot string        `json:"storageRoot"`
	CodeHash    string        `json:"codeHash"`
}

type SendTxResult struct {
	TXhash string `json:"TXhash,omitempty"`
	ErrMsg string `json:"ErrMsg,omitempty"`
}
type RpcMakeVccProofJson struct {
	Proof  map[string]interface{} `json:"proof,omitempty"`
	ErrMsg string                 `json:"ErrMsg,omitempty"`
}

type MakeCCCExistenceProofJson struct {
	Proof  map[string]interface{} `json:"proof,omitempty"`
	ErrMsg string                 `json:"ErrMsg,omitempty"`
}

//GetCCCRelativeTx
type GetCCCRelativeTxJson struct {
	Proof  map[string]interface{} `json:"proof,omitempty"`
	ErrMsg string                 `json:"ErrMsg,omitempty"`
}
type CompileContractJson struct {
	Test   map[string]interface{} `json:"test,omitempty"`
	ErrMsg string                 `json:"ErrMsg,omitempty"`
}

type TransactionResult struct {
	ChainId   int      `json:"chainId"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	Nonce     int      `json:"nonce"`
	Value     *big.Int `json:"value"`
	Input     string   `json:"input"`
	Hash      string   `json:"hash"`
	UseLocal  bool     `json:"uselocal"`
	Extra     string   `json:"extra"` // 目前用来存交易类型，不存在时为普通交易，否则会对应特殊操作
	Timestamp uint64   `json:"timestamp"`
}

type TxResult struct {
	Transaction     TransactionResult `json:"tx"`
	Root            string            `json:"root"`
	Status          int               `json:"status"`
	Logs            interface{}       `json:"logs"`
	TransactionHash string            `json:"transactionHash"`
	ContractAddress string            `json:"contractAddress"`
	Out             string            `json:"out"`
	BlockHeight     int               `json:"blockHeight"`
	ErrMsg          string            `json:"ErrMsg,omitempty"`
}

// type TxResultHash struct {
// 	Tx              TransactionResult `json:"tx"`
// 	Root            string      `json:"root"`
// 	Status          int         `json:"status"`
// 	Logs            interface{}      `json:"logs"`
// 	TransactionHash string      `json:"transactionHash"`
// 	ContractAddress string      `json:"contractAddress"`
// 	Out             string      `json:"out"`
// 	BlockHeight     int         `json:"blockHeight"`
// }
type GetBlockResult struct {
	Hash          string `json:"hash"`          // 此块的hsh
	Previoushash  string `json:"previousHash"`  // 父块的hash
	ChainId       int    `json:"chainId"`       //
	Height        int    `json:"height"`        // 查询块的块高
	Empty         bool   `json:"empty"'`        // 是否是空块
	RewardAddress string `json:"rewardaddress"` // 接收地址
	Mergeroot     string `json:"mergeRoot"`     // 合并其他链转块数据hash
	Deltaroot     string `json:"deltaRoot"`     // 跨链转账数据hash
	Stateroot     string `json:"stateRoot"`     // 状态hash
	RREra         string `json:"rrera"`
	RRCurrent     string `json:"rrcurrent"`
	RRNext        string `json:"rrnext"`
	Txcount       int    `json:"txcount"`
	Timestamp     int64  `json:"timestamp"`
	ErrMsg        string `json:"ErrMsg,Omitempty"`
}

type GetChainInfo struct {
	ChainId      int    `json:"chainId"`
	DataNodeId   string `json:"dataNodeId"`
	DataNodeIp   string `json:"dataNodeIp"`
	DataNodePort int    `json:"dataNodePort"`
	Mode         int    `json:"mode"`
	Parent       int    `json:"parent"`
	ErrMsg       string `json:"ErrMsg,Omitempty"`
}

/*
"chainId": 2,
"from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
"to": "0x0000000000000000000000000000000000020000",
"nonce": 0,
"value": 0,
"input": "0x000000022c7536e3605d9c16a7a3d7b1898e529396a65c230000000000000000000000034fa1c4e6182b6b7f3bca273390cf587b50b4731100000000000456440101",
"hash": "0x0ea5dad47833fc6286357b6bd6c1a4e910def5f4432a1a59bde0f816c3dd18e0",
"timestamp": 1560425588
*/
type GetTransactions struct {
	ChainId   int    `json:"chainId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Nonce     int    `json:"nonce"`
	Value     int    `json:"value"`
	Input     string `json:"input"`
	Hash      string `json:"hash"`
	Timestamp int64  `json:"timestamp"`
}

type GetChainStats struct {
	ChainId       int `json:"chainId"`
	Currentheight int `json:"currentheight"`
}

type GetCommittee struct {
	ChainId       int32    `json:"chainId"`
	MemberDetails []string `json:"memberDetails"`
	Epoch         int      `json:"epoch"`
	ErrMsg        string   `json:"ErrMsg,Omitempty"`
}

type GetMultiStatsResult struct {
	ErrMsg string `json:"ErrMsg,Omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (pointer *RequestResult) ToStringArray() ([]string, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.([]interface{})

	new := make([]string, len(result))
	for i, v := range result {
		new[i] = v.(string)
	}

	return new, nil

}

func (pointer *RequestResult) ToComplexString() (types.ComplexString, error) {

	if err := pointer.checkResponse(); err != nil {
		return "", err
	}

	result := (pointer).Result.(interface{})

	return types.ComplexString(result.(string)), nil

}

func (pointer *RequestResult) ToString() (string, error) {

	if err := pointer.checkResponse(); err != nil {
		return "", err
	}

	result := (pointer).Result.(interface{})

	return result.(string), nil

}

func (pointer *RequestResult) ToInt() (int64, error) {

	if err := pointer.checkResponse(); err != nil {
		return 0, err
	}

	result := (pointer).Result.(interface{})

	hex := result.(string)

	numericResult, err := strconv.ParseInt(hex, 16, 64)

	return numericResult, err

}

func (pointer *RequestResult) ToBigInt() (*big.Int, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	res := (pointer).Result.(interface{})

	ret, success := big.NewInt(0).SetString(res.(string)[2:], 16)

	if !success {
		return nil, errors.New(fmt.Sprintf("Failed to convert %s to BigInt", res.(string)))
	}

	return ret, nil
}

func (pointer *RequestResult) ToComplexIntResponse() (types.ComplexIntResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return types.ComplexIntResponse(0), err
	}

	result := (pointer).Result.(interface{})

	var hex string

	switch v := result.(type) {
	// Testrpc returns a float64
	case float64:
		hex = strconv.FormatFloat(v, 'E', 16, 64)
		break
	default:
		hex = result.(string)
	}

	cleaned := strings.TrimPrefix(hex, "0x")

	return types.ComplexIntResponse(cleaned), nil

}

func (pointer *RequestResult) ToBoolean() (bool, error) {

	if err := pointer.checkResponse(); err != nil {
		return false, err
	}

	result := (pointer).Result.(interface{})

	return result.(bool), nil

}

func (pointer *RequestResult) ToSignTransactionResponse() (*SignTransactionResponse, error) {
	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	signTransactionResponse := &SignTransactionResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), signTransactionResponse)

	return signTransactionResponse, err
}

func (pointer *RequestResult) ToTransactionResponse() (*TransactionResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	transactionResponse := &TransactionResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), transactionResponse)

	return transactionResponse, err

}

func (pointer *RequestResult) ToTransactionReceipt() (*TransactionReceipt, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	transactionReceipt := &TransactionReceipt{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), transactionReceipt)

	return transactionReceipt, err

}

func (pointer *RequestResult) ToBlock() (*Block, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	block := &Block{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), block)

	return block, err

}

func (pointer *RequestResult) ToSyncingResponse() (*SyncingResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	var result map[string]interface{}

	switch (pointer).Result.(type) {
	case bool:
		return &SyncingResponse{}, nil
	case map[string]interface{}:
		result = (pointer).Result.(map[string]interface{})
	default:
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	syncingResponse := &SyncingResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	json.Unmarshal([]byte(marshal), syncingResponse)

	return syncingResponse, nil

}

// To avoid a conversion of a nil interface
func (pointer *RequestResult) checkResponse() error {

	if pointer.Error != nil {
		return errors.New(pointer.Error.Message)
	}

	if pointer.Result == nil {
		return customerror.EMPTYRESPONSE
	}

	return nil

}
