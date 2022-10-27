package crust

import (
	"github.com/shopspring/decimal"
)

type Header struct {
	Digest         Digest `json:"digest"`
	ExtrinsicsRoot string `json:"extrinsicsRoot"`
	Number         string `json:"Number"`
	ParentHash     string `json:"parentHash"`
	stateRoot      string `json:"stateRoot"`
	Height         int64  `json:"height"` //Number转成height
}
type Digest struct {
	Logs []string `json:"Logs"`
}
type Transaction struct {
	Txid     string          `json:"txid"`
	From     string          `json:"from"`
	To       string          `json:"to"`
	Value    decimal.Decimal `json:"value"`
	Fee      decimal.Decimal `json:"fee"`
	Function string          `json:"function"`
}

type Block struct {
	Hash          string      `json:"hash"`
	Block         BlockDetail `json:"block"`
	Justification interface{} `json:"justification"`
}
type BlockDetail struct {
	Extrinsics []string `json:"extrinsics"`
	Header     Header   `json:"header"`
}
