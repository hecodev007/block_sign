package atom

import (
	"atomDataServer/utils"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/shopspring/decimal"
	"github.com/tendermint/tendermint/types"
	"log"
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
	BlockId BlockID `json:"block_id"`
	Block struct {
		Header ResponseBlockHeader `json:"header"`
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
	Error string `json:"error"`
	//Hash         string    `json:"hash"`
	//ParentHash   string    `json:"parent_hash"`
	//Height       int64     `json:"height"`
	//ChainID      string    `json:"chain_id"`
	//Timestamp    time.Time `json:"timestamp"`
	//Transactions []string  `json:"transactions"`
}

// BlockID defines the unique ID of a block as its Hash and its PartSetHeader
type BlockID struct {
	Hash        string `json:"hash"`
	PartsHeader struct {
		Total int64 `json:"total"`
		Hash  string `json:"hash"`
	} `json:"parts"`
}

type ResponseBlockHeader struct {
	// basic block info
	ChainID  string    `json:"chain_id"`
	Height   decimal.Decimal    `json:"height"`
	Time     time.Time `json:"time"`
	NumTxs   string    `json:"num_txs"`
	//TotalTxs string    `json:"total_txs"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	//LastCommitHash string `json:"last_commit_hash"` // commit from validators from the last block
	//DataHash       string `json:"data_hash"`        // transactions
	//
	//// hashes from the app output from the prev block
	//ValidatorsHash     string `json:"validators_hash"`      // validators for the current block
	//NextValidatorsHash string `json:"next_validators_hash"` // validators for the next block
	//ConsensusHash      string `json:"consensus_hash"`       // consensus params for current block
	//AppHash            string `json:"app_hash"`             // state after txs from the previous block
	//LastResultsHash    string `json:"last_results_hash"`    // root hash of all results from the txs from the previous block
	//
	//// consensus info
	//EvidenceHash    string `json:"evidence_hash"`    // evidence included in the block
	//ProposerAddress string `json:"proposer_address"` // original proposer of the block
}

// Single block (with meta)
type ResponseBlock struct {
	BlockID BlockID             `json:"block_id"` // the block hash and partsethash
	Block struct {
		Header ResponseBlockHeader `json:"header"`
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
}

func (proxy *ResponseBlock) toBlock() *Block {


	block := Block{
		BlockId:  proxy.BlockID,
		Block:proxy.Block,
	}
	return &block
}

type ResponseTx struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
	RawLog string `json:"raw_log,omitempty"`
	Logs   []struct {
		Events []struct{
			Type string `json:"type"`
			Attributes []struct{
				Key string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
		//MsgIndex int64  `json:"msg_index"`
		//Success  bool   `json:"success"`
		//Log      string `json:"log"`
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

type MsgDelegate struct {
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorAddress string    `json:"validator_address"`
	Amount           sdk.Coin  `json:"amount"`
	UnlockTime       time.Time `json:"unlock_time"`
}

func (proxy *ResponseTx) ToTx() (tx *Transaction, err error) {
	tx = &Transaction{
		Hash: proxy.TxHash,
		//RawLogs: proxy.RawLog,
		Memo: proxy.Tx.Value.Memo,
	}
	tx.Timestamp, _ = time.Parse(time.RFC3339, proxy.Timestamp)
	tx.GasWanted, _ = utils.ParseInt64(proxy.GasWanted)
	tx.GasUsed, _ = utils.ParseInt64(proxy.GasUsed)
	tx.BlockHeight, _ = utils.ParseInt64(proxy.Height)

	if proxy.Tx.Type == "cosmos-sdk/StdTx" {
		Fee, err := AtomToInt(proxy.Tx.Value.Fee.Amount.String())
		if err != nil {
			panic(proxy.TxHash + err.Error()+"  "+proxy.Tx.Value.Fee.Amount.String())
		}
		tx.Fee = Fee
		tx.Memo = proxy.Tx.Value.Memo
		msgs := proxy.Tx.Value.Msgs
		//log.Printf("get msg num : %d ", len(msgs))
		num := 0
		if len(msgs) > 0 {
			for i, tmp := range msgs {
				//log.Printf("get msg type : %s ", tmp.Type)
				valuebyte, _ := json.Marshal(tmp.Value)
				switch tmp.Type {
				case TypeMsgDelegate:
					break
				case "cosmos-sdk/MsgSend":
					if num > 0 {
						//2个send:D828A6B5506458D9BF11FBFFAE659B3CEA049DA5246963D450F42105284AAB0C
						panic(proxy.TxHash)
					}
					num++

					tx.Type = "send"
					if len(proxy.Logs) > i && len(proxy.Logs[i].Events) == 4 && proxy.Logs[i].Events[2].Type == "message" && proxy.Logs[i].Events[3].Type == "transfer" {
						tx.Success = true
					} else {
						//panic(proxy.TxHash)
					}

					var msg MsgSend
					err := json.Unmarshal(valuebyte, &msg)
					if err != nil {
						log.Printf("msgsend value : %T , %v ", tmp.Value, tmp.Value)
						return nil, err
					}
					tx.From = msg.FromAddress
					tx.To = msg.ToAddress
					Value, err := AtomToInt(msg.Amount.String())
					if err != nil {
						//panic(proxy.TxHash + " " + err.Error())
						Value = 0
					}
					tx.Value = Value
					tx.UnlockTime = msg.UnlockTime

					return tx, nil
				case TypeMultiSend:
					break
				case TypeMsgDeposit:
					break
				case TypeMsgWithdrawDelegationReward:
					break
				default:
					break
				}
			}
		}
	} else {
		//return nil, fmt.Errorf("don't support tx type : %s", tx.Type)
		return tx,nil
	}

	return tx, nil
}

type Transaction struct {
	Hash        string `json:"hash"`
	BlockHeight int64  `json:"block_height"`
	//BlockHash   string    `json:"block_hash"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Value      int64     `json:"value"`
	Fee        int64     `json:"fee,omitempty"`
	Memo       string    `json:"memo,omitempty"`
	Success    bool      `json:"success"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	UnlockTime string    `json:"unlock_time"`

	GasUsed   int64   `json:"gas_used,omitempty" `
	GasWanted int64   `json:"gas_wanted,omitempty"`
	GasPrice  string  `json:"gas_price,omitempty"`
	RawLogs   string  `json:"raw_logs,omitempty"`
	TxMsgs    []TxMsg `json:"tx_msgs"`
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
	MsgIndex int64    `json:"msg_index"`
	Success  bool     `json:"success"`
	Log      string   `json:"log"`
	Events   []*Event `json:"events"`
}
type Event struct {
	Type       string       `json:"type"`
	Attributes []*Attribute `json:"attributes"`
}
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
