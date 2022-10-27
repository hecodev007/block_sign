package cph

type CphBlockStruct struct {
	BlockType    int      `json:"BlockType"`
	Exceptions   string   `json:"exceptions"`
	Hash         string   `json:"hash"`
	KeyHash      string   `json:"keyHash"`
	Number       string   `json:"number"`
	ParentHash   string   `json:"parentHash"`
	Signature    string   `json:"signature"`
	Timestamp    string   `json:"timestamp"`
	Transactions []string `json:"transactions"`
}

type CphTransactionStruct struct {
	Version          string `json:"version"`
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	From             string `json:"from"`
}
type CphTransactionReceipt struct {
	ContractAddress string   `json:"contractAddress"`
	Logs            []string `json:"logs"`
	From            string   `json:"from"`
	To              string   `json:"to"`
	Status          string   `json:"status,omitempty"`
}
