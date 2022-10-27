package wit

import "encoding/json"

type Params struct {
	Id      int64       `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type Params2 struct {
	Id      string      `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}
type Response struct {
	Id      interface{}     `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type Block struct {
	Height     int64  `json:"height"`
	Hash       string `json:"blockhash"`
	Txns       *Txns  `json:"txns"`
	TxnsHashes struct {
		ValueTransfer []string `json:"value_transfer"`
	} `json:"txns_hashes"`
}
type Txns struct {
	ValueTransferTxns []*ValueTransferTxn `json:"value_transfer_txns"`
}
type ValueTransferTxn struct {
	Body *Body `json:"body"`
}
type Body struct {
	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`
}
type Output struct {
	Address  string `json:"pkh"`
	TimeLock int64  `json:"time_lock"`
	Value    int64  `json:"value"`
}

type Input struct {
	OutputPointer string `json:"output_pointer"`
}

type Transaction struct {
	BlockHash   string `json:"blockHash"`
	Transaction struct {
		ValueTransferTxn *ValueTransferTxn `json:"ValueTransfer"`
		Mint             *Mint             `json:"mint"`
	} `json:"transaction"`
}

type Mint struct {
	Epoch   int64     `json:"epoch"`
	Outputs []*Output `json:"outputs"`
}
