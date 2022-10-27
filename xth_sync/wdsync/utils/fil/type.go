package fil

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

type ChainHead struct {
	//Cids      []string
	Cids   []map[string]string `json:"Cids"`
	Blocks []*BlockHeader      `json:"Blocks"`
	Height int64               `json:"Height"`
}

type Transaction struct {
	Version    int64           `json:"version"`
	To         string          `json:"To"`
	From       string          `json:"From"`
	Nonce      int64           `json:"Nonce"`
	Value      decimal.Decimal `json:"Value"`
	GasLimit   int64           `json:"GasLimit"`
	GasFeeCap  decimal.Decimal `json:"GasFeeCap"`
	GasPremium decimal.Decimal `json:"GasPremium"`
	Fee        decimal.Decimal `json:"fee"`
	Method     int64           `json:"Method"`
	Params     string          `json:"Params"`
	//Cid        string          `json:"cid"`
	Cid     map[string]string `json:"Cid"`

}
type BlockMessages struct {
	BlsMessages   []*Transaction      `json:"BlsMessages"`
	SecpkMessages []*SignedMessage    `json:"SecpkMessages"`
	Cids          []map[string]string `json:"Cids"`
}
type Message struct {
	Cid     map[string]string `json:"Cid"`
	Message *Transaction      `json:"Message"`
}
type SignedMessage struct {
	Message   *Transaction
	Signature interface{}
}
type Receipt struct {
	ExitCode int64
	Return   string
	GasUsed  int64
}

func (r *Receipt) Success() bool {
	return r.ExitCode == 0 && r.GasUsed != 0 && r.Return != ""
}

type BlockHeader struct {
	Miner         string      `json:"Miner"`
	Ticket        interface{} `json:"Ticket"`
	ElectionProof interface{} `json:"ElectionProof"`
	BeaconEntries interface{} `json:"BeaconEntries"`
	WinPoStProof  interface{} `json:"WinPoStProof"`
	//Parents               []string
	Parents               []map[string]string `json:"Parents"`
	ParentWeight          string              `json:"ParentWeight"`
	Height                int64               `json:"Height"`
	ParentStateRoot       map[string]string   `json:"ParentStateRoot"`
	ParentMessageReceipts map[string]string   `json:"ParentMessageReceipts"`
	Messages              map[string]string   `json:"Messages"`
	BLSAggregate          interface{}         `json:"BLSAggregate"`
	Timestamp             uint64              `json:"Timestamp"`
	BlockSig              interface{}         `json:"BlockSig"`
	ForkSignaling         uint64              `json:"ForkSignaling"`
	ParentBaseFee         string              `json:"ParentBaseFee"`
	validated             bool                `json:"validated"`
	Cid                   string              `json:"Cid"`
}
type BlockTransactions struct {
	BlsMessages   []Transaction   `json:"SecpkMessages"`
	SecpkMessages []SignedMessage `json:"SecpkMessages"`
	Cids          []string        `json:"Cids"`
}
type SyncState struct {
	ActiveSyncs []*ActiveSync `json:"ActiveSyncs"`
}
func (s *SyncState)String()string{
	str,_:=json.Marshal(s)
	return string(str)
}
type ActiveSync struct {
	Height int64   `json:"Height"`
	Base   *Base   `json:"Base"`
	Target *Target `json:"Target"`
}
type Target struct {
	Cids   []map[string]string `json:"Cids"`
	Blocks []*BlockHeader      `json:"Blocks"`
	Height int64 `json:"height"`
}
type Base struct {
	Cids   []map[string]string `json:"Cids"`
	Blocks []*BlockHeader      `json:"Blocks"`
}

//type Block struct {
//	Miner   string              `json:"Miner"`
//	Parents []map[string]string `json:"Parents"`
//}
