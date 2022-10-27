package bifrost

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/shopspring/decimal"
)

type Method struct {
	Call types.Call
	Args types.Args
}

type Args struct {
	To     types.MultiAddress
	Amount types.UCompact
}

type BlockExtrinsicParams struct {
	from, to, sig, era, txid, fee string
	nonce                         int64
	extrinsicIdx, length          int
}

type EventResult struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
	ExtrinsicIdx int    `json:"extrinsic_idx"`
	EventIdx     int    `json:"event_idx"`
	Status       string `json:"status"`
	Weight       int64  `json:"weight"` //权重
}

type NodeBlock struct {
	Block struct {
		Extrinsics []string `json:"extrinsics"`
	} `json:"block"`
}

type BlockResponse struct {
	Height     int64                `json:"height"`
	ParentHash string               `json:"parent_hash"`
	BlockHash  string               `json:"block_hash"`
	Timestamp  int64                `json:"timestamp"`
	Extrinsic  []*ExtrinsicResponse `json:"extrinsic"`
}

type ExtrinsicResponse struct {
	Type            string `json:"type"`   //Transfer or another
	Status          string `json:"status"` //success or fail
	Txid            string `json:"txid"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	Fee             string `json:"fee"`
	Signature       string `json:"signature"`
	Nonce           int64  `json:"nonce"`
	Era             string `json:"era"`
	ExtrinsicIndex  int    `json:"extrinsic_index"`
	EventIndex      int    `json:"event_index"`
	ExtrinsicLength int    `json:"extrinsic_length"`
}

type QueryInfo struct {
	Class      string          `json:"class"`
	PartialFee decimal.Decimal `json:"partialFee"`
	Weight     int64           `json:"weight"`
}
