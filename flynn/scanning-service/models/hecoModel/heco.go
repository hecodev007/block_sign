package hecoModel

// Block - block object
type Block struct {
	Number           string        `json:"number"`
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parentHash"`
	Nonce            string        `json:"nonce"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	LogsBloom        string        `json:"logsBloom"`
	TransactionsRoot string        `json:"transactionsRoot"`
	StateRoot        string        `json:"stateRoot"`
	Miner            string        `json:"miner"`
	Difficulty       string        `json:"difficulty"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	ExtraData        string        `json:"extraData"`
	Size             string        `json:"size"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Timestamp        string        `json:"timestamp"`
	Uncles           []string      `json:"uncles"`
	Transactions     []interface{} `json:"transactions"`
}

// Transaction - transaction object
type Transaction struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	Hash        string `json:"hash"`
	Input       string `json:"input"`
	Nonce       string `json:"nonce"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

type TransactionReceipt struct {
	ContractAddress   string           `json:"contractAddress"`
	Logs              []TransactionLog `json:"logs"`
	From              string           `json:"from"`
	To                string           `json:"to"`
	Status            string           `json:"status,omitempty"`
	TransactionHash   string           `json:"transactionHash"`
	TransactionIndex  string           `json:"transactionIndex"`
	GasUsed           string           `json:"gasUsed"`
	CumulativeGasUsed string           `json:"cumulativeGasUsed"`
}

type TransactionLog struct {
	Address          string   `json:"address"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	Data             string   `json:"data"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
	Topics           []string `json:"topics"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
}
