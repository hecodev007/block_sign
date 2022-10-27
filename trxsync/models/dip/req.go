package dip

import (
	sdk "github.com/Dipper-Labs/Dipper-Protocol/types"
	"github.com/tendermint/tendermint/types"
	"time"
)

// BlockID defines the unique ID of a block as its Hash and its PartSetHeader
type BlockID struct {
	Hash        string `json:"hash"`
	PartsHeader struct {
		Total string `json:"total"`
		Hash  string `json:"hash"`
	} `json:"parts"`
}

type ResponseBlockHeader struct {
	// basic block info
	ChainID  string    `json:"chain_id"`
	Height   string    `json:"height"`
	Time     time.Time `json:"time"`
	NumTxs   string    `json:"num_txs"`
	TotalTxs string    `json:"total_txs"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash string `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       string `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     string `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash string `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      string `json:"consensus_hash"`       // consensus params for current block
	AppHash            string `json:"app_hash"`             // state after txs from the previous block
	LastResultsHash    string `json:"last_results_hash"`    // root hash of all results from the txs from the previous block

	// consensus info
	EvidenceHash    string `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress string `json:"proposer_address"` // original proposer of the block
}
type ResponseBlock struct {
	BlockMeta struct {
		BlockID BlockID             `json:"block_id"` // the block hash and partsethash
		Header  ResponseBlockHeader `json:"header"`   // The block's Header
	} `json:"block_meta"`
	Block struct {
		Herder ResponseBlockHeader `json:"header"`
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
}

type ResponseTx struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
	RawLog string `json:"raw_log,omitempty"`
	Logs   []struct {
		MsgIndex int64  `json:"msg_index"`
		Success  bool   `json:"success"`
		Log      string `json:"log"`
	} `json:"logs,omitempty"`
	GasWanted string `json:"gas_wanted,omitempty"`
	GasUsed   string `json:"gas_used,omitempty"`
	Tx        struct {
		Type  string `json:"type"`
		Value struct {
			Msgs []struct {
				Type  string      `json:"type"`
				Value interface{} `json:"value"`
			} `json:"msg"`
			Fee struct {
				Amount sdk.Coins `json:"amount"`
				Gas    string    `json:"gas"`
			} `json:"fee"`
			Memo string `json:"memo"`
		} `json:"value"`
	} `json:"tx"`
	Timestamp string `json:"timestamp"`
	Error     string `json:"error"`
}

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      sdk.Coins `json:"amount"`
	UnlockTime  string    `json:"unlock_time"`
}
