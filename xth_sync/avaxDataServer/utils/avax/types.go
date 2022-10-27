package avax

import (
	"fmt"
	"strings"
	"time"
)

//区块链详情
type BlockChainInfo struct {
	Chain         string  `json:"chain"`
	Blocks        uint    `json:"blocks"`
	Headers       uint    `json:"headers"`
	Bestblockhash string  `json:"bestblockhash"`
	Difficulty    float64 `json:"difficulty"`
	Mediantime    uint64  `json:"mediantime"`
	Chainwork     string  `json:"chainwork"`
}

//区块链详情
type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	Size              int64    `json:"size"`
	Height            int64    `json:"height"`
	Version           int      `json:"version"`
	Time              int64    `json:"time"`
	Chainwork         string   `json:"chainwork"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     string   `json:"nextblockhash"`
	Txs               []string `json:"tx"`
}

//区块链详情
type BlockWithTx struct {
	Hash              string         `json:"hash"`
	Confirmations     int64          `json:"confirmations"`
	Size              int64          `json:"size"`
	Height            int64          `json:"height"`
	Version           int            `json:"version"`
	Time              int64          `json:"time"`
	Chainwork         string         `json:"chainwork"`
	PreviousBlockHash string         `json:"previousblockhash"`
	NextBlockHash     string         `json:"nextblockhash"`
	Txs               []*Transaction `json:"tx"`
}

type Transactionbtc struct {
	Txid          string       `json:"txid"`
	Hash          string       `json:"hash"`
	Version       int          `json:"version"`
	Size          int          `json:"size"`
	LockTime      int64        `json:"locktime"`
	Vin           []proxyTxIn  `json:"vin"`
	Vout          []proxyTxOut `json:"vout"`
	BlockHash     string       `json:"blockhash"`
	Confirmations int64        `json:"confirmations"`
	Time          int64        `json:"time"`
}

type proxyTxIn struct {
	Txid     string `json:"txid,omitempty"`
	Vout     int    `json:"vout,omitempty"`
	Coinbase string `json:"coinbase,omitempty"`
	Sequence int64  `json:"sequence"`
}

type proxyTxOut struct {
	Value        float64      `json:"value"`
	Index        int          `json:"n"`
	ScriptPubkey scriptPubkey `json:"scriptPubkey"`
}
type scriptPubkey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

func (sp *scriptPubkey) GetAddress() ([]string, error) {

	switch sp.Type {
	case "":
		return sp.Addresses, nil
	default:
		return nil, fmt.Errorf("don't support tx %s", sp.Type)
	}
}

type ListTransaction struct {
	Count        int64          `json:"count"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct {
	ID      string `json:"id"`
	ChainID string `json:"chainID"`
	Type    string `json:"type"`

	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`

	InputTotals         map[string]string `json:"inputTotals"`
	OutputTotals        map[string]string `json:"outputTotals"`
	ReusedAddressTotals map[string]string `json:"reusedAddressTotals"`

	CanonicalSerialization []byte    `json:"canonicalSerialization,omitempty"`
	Timestamp              time.Time `json:"timestamp"`
}
type Input struct {
	Output *Output            `json:"output"`
	Creds  []InputCredentials `json:"credentials"`
}
type InputCredentials struct {
	Address   Address `json:"address"`
	PublicKey []byte  `json:"public_key"`
	Signature []byte  `json:"signature"`
}
type Output struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transactionID"`
	OutputIndex   uint64    `json:"outputIndex"`
	AssetID       string    `json:"assetID"`
	OutputType    uint32    `json:"outputType"`
	Amount        string    `json:"amount"`
	Locktime      uint64    `json:"locktime"`
	Threshold     uint64    `json:"threshold"`
	Addresses     []Address `json:"addresses"`
	CreatedAt     time.Time `json:"timestamp"`

	RedeemingTransactionID string `json:"redeemingTransactionID"`
}

type Address string

func (Addr *Address) UnmarshalJSON(b []byte) error {
	Bstring := string(b)
	Bstring = strings.ReplaceAll(Bstring, "\"", "")
	if strings.Index(Bstring, "-") >= 0 {
		*Addr = Address(Bstring)
	} else {
		*Addr = Address("X-" + Bstring)
	}
	return nil
}
