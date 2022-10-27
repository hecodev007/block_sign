package atom

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/shopspring/decimal"
	"github.com/tendermint/tendermint/types"
	"telosDataServer/common/log"
	"telosDataServer/utils"
	"time"
)

const (
	TypeMsgSend                     = "send"
	TypeMsgDelegate                 = "delegate"
	TypeMultiSend                   = "multisend"
	TypeMsgDeposit                  = "deposit"
	TypeMsgWithdrawDelegationReward = "withdraw_delegator_reward"
)

type Block struct {
	Hash         string    `json:"hash"`
	ParentHash   string    `json:"parent_hash"`
	Height       int64     `json:"height"`
	ChainID      string    `json:"chain_id"`
	Timestamp    time.Time `json:"timestamp"`
	Transactions []string  `json:"transactions"`
}

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

// Single block (with meta)
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

func (proxy *ResponseBlock) toBlock() *Block {
	block := Block{
		Hash:       proxy.BlockMeta.BlockID.Hash,
		ParentHash: proxy.Block.Herder.LastBlockID.Hash,
		ChainID:    proxy.Block.Herder.ChainID,
		Timestamp:  proxy.Block.Herder.Time,
	}
	block.Height, _ = utils.ParseInt64(proxy.Block.Herder.Height)
	for _, tx := range proxy.Block.Data.Txs {
		block.Transactions = append(block.Transactions, fmt.Sprintf("%X", tx.Hash()))
	}
	return &block
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
}

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      sdk.Coins `json:"amount"`
	UnlockTime  string    `json:"unlock_time"`
}

type MsgDelegate struct {
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorAddress string    `json:"validator_address"`
	Amount           sdk.Coin  `json:"amount"`
	UnlockTime       time.Time `json:"unlock_time"`
}

func (proxy *ResponseTx) toCetTx() (*Transaction, error) {
	tx := &Transaction{
		Hash:    proxy.TxHash,
		RawLogs: proxy.RawLog,
		Type:    proxy.Tx.Type,
	}
	tx.Timestamp, _ = time.Parse(time.RFC3339, proxy.Timestamp)
	tx.GasWanted, _ = utils.ParseInt64(proxy.GasWanted)
	tx.GasUsed, _ = utils.ParseInt64(proxy.GasUsed)
	tx.BlockHeight, _ = utils.ParseInt64(proxy.Height)

	if tx.Type == "cosmos-sdk/StdTx" {
		tx.Fee = proxy.Tx.Value.Fee.Amount.String()
		tx.Memo = proxy.Tx.Value.Memo

		msgs := proxy.Tx.Value.Msgs
		log.Infof("get msg num : %d ", len(msgs))
		if len(msgs) > 0 {
			for i, tmp := range msgs {
				log.Infof("get msg type : %s ", tmp.Type)
				valuebyte, _ := json.Marshal(tmp.Value)
				switch tmp.Type {
				case TypeMsgDelegate:
					txmsg := TxMsg{
						Index: i,
						Type:  tmp.Type,
					}
					if len(proxy.Logs) > i {
						txmsg.Log = proxy.Logs[i].Log
						txmsg.Success = proxy.Logs[i].Success
					}

					var msg MsgDelegate
					if err := json.Unmarshal(valuebyte, &msg); err == nil {
						txmsg.From = msg.DelegatorAddress
						txmsg.To = msg.ValidatorAddress
						txmsg.Amount = msg.Amount.String()
					} else {
						log.Infof("msgsend value : %T , %v ", tmp.Value, tmp.Value)
					}
					tx.TxMsgs = append(tx.TxMsgs, txmsg)
					break
				case "bankx/MsgSend":
					txmsg := TxMsg{
						Index: i,
						Type:  tmp.Type,
					}
					if len(proxy.Logs) > i {
						txmsg.Log = proxy.Logs[i].Log
						txmsg.Success = proxy.Logs[i].Success
					}

					var msg MsgSend
					if err := json.Unmarshal(valuebyte, &msg); err == nil {
						txmsg.From = msg.FromAddress
						txmsg.To = msg.ToAddress
						txmsg.Amount = msg.Amount.String()
						txmsg.UnlockTime = msg.UnlockTime
					} else {
						log.Infof("msgsend value : %T , %v ", tmp.Value, tmp.Value)
					}
					tx.TxMsgs = append(tx.TxMsgs, txmsg)
					break
				case TypeMultiSend:
					break
				case TypeMsgDeposit:
					break
				case TypeMsgWithdrawDelegationReward:
					break
				}
			}
		}
	} else {
		return nil, fmt.Errorf("don't support tx type : %s", tx.Type)
	}

	return tx, nil
}

type Transaction struct {
	Hash        string    `json:"hash"`
	BlockHeight int64     `json:"block_height"`
	GasWanted   int64     `json:"gas_wanted,omitempty"`
	GasUsed     int64     `json:"gas_used,omitempty" `
	GasPrice    string    `json:"gas_price,omitempty"`
	RawLogs     string    `json:"raw_logs,omitempty"`
	Type        string    `json:"type"`
	Fee         string    `json:"fee,omitempty"`
	Memo        string    `json:"memo,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	TxMsgs      []TxMsg   `json:"tx_msgs"`
}

type TxMsg struct {
	Index      int    `json:"index"`
	Type       string `json:"type"`
	Success    bool   `json:"success"`
	Log        string `json:"log"`
	From       string `json:"from"`
	To         string `json:"to"`
	Amount     string `json:"amount"`
	UnlockTime string `json:"unlock_time"`
}

type MessageLog struct {
	MsgIndex string `json:"msg_index"`
	Success  bool   `json:"success"`
	Log      string `json:"log"`
}

type ProxyTx struct {
	Type  string     `json:"type"`
	Value auth.StdTx `json:"value"`
}

type ProxyTransaction struct {
	Height    string        `json:"height"`
	TxHash    string        `json:"txhash"`
	RawLog    string        `json:"raw_log,omitempty"`
	Logs      []*MessageLog `json:"logs,omitempty"`
	GasWanted string        `json:"gas_wanted,omitempty"`
	GasUsed   string        `json:"gas_used,omitempty"`
	Tx        ProxyTx       `json:"tx"`
	Timestamp string        `json:"timestamp"`
}

func GetCoinNum(coin, Denom string) decimal.Decimal {
	feecoins, err := sdk.ParseCoins(coin)
	if err != nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(feecoins.AmountOf(Denom).BigInt(), 0)
}
