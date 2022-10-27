package stx

import "github.com/shopspring/decimal"

type BlocksResult struct {
	Error   string `json:"error"`
	Limit   int64  `json:"limit"`
	Offset  int64  `json:"offset"`
	Total   int64  `json:"total"`
	Results []*Block
}

type Block struct {
	Height    int64    `json:"height"`
	Hash      string   `json:"hash"`
	Txs       []string `json:"txs"`
	Canonical bool     `json:"canonical"`
}

type Transaction struct {
	Error             string          `json:"error"`
	TxId              string          `json:"tx_id"`
	TxType            string          `json:"tx_type"`
	Nonce             int64           `json:"nonce"`
	FeeRate           decimal.Decimal `json:"fee_rate"`
	SenderAddress     string          `json:"sender_address"`
	Sponsored         bool            `json:"sponsored"`
	PostConditionMode string          `json:"post_condition_mode"`
	TxStatus          string          `json:"tx_status"` //success
	BlockHash         string          `json:"block_hash"`
	BlockHeight       int64           `json:"block_height"`
	TokenTransfer     struct {
		RecipientAddress string          `json:"recipient_address"`
		Amount           decimal.Decimal `json:"amount"`
		Memo             string          `json:"memo"`
	} `json:"token_transfer"`
	ContractCall struct {
		ContractId   string `json:"contract_id"`
		FunctionName string `json:"function_name"`
		FunctionArgs struct {
			Name string `json:"name"`
		} `json:"function_args"`
	} `json:"contract_call"`
	Events []struct {
		EventIndex int    `json:"event_index"`
		EventType  string `json:"event_type"`
		Asset      struct {
			AssetEventType string `json:"asset_event_type"`
		} `json:"asset"`
	}
}
