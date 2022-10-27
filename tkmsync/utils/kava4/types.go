package kava4

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/types"
	"rsksync/utils"
	"time"
)

const (
	TypeMsgSend                     = "send"
	TypeMsgDelegate                 = "delegate"
	TypeMultiSend                   = "multisend"
	TypeMsgDeposit                  = "deposit"
	TypeMsgWithdrawDelegationReward = "withdraw_delegator_reward"

	MainnetDenom = "ukava"
	MainChainID  = "kava-4"
)

type Block struct {
	Hash         string    `json:"hash"`
	ParentHash   string    `json:"parent_hash"`
	Height       int64     `json:"height"`
	ChainID      string    `json:"chain_id"`
	Timestamp    time.Time `json:"timestamp"`
	Transactions []string  `json:"transactions"`
}
type Transaction struct {
	Hash        string    `json:"hash"`
	BlockHeight int64     `json:"block_height"`
	GasWanted   int64     `json:"gas_wanted,omitempty"`
	GasUsed     int64     `json:"gas_used,omitempty" `
	GasPrice    uint64    `json:"gas_price,omitempty"`
	RawLogs     string    `json:"raw_logs,omitempty"`
	Type        string    `json:"type"`
	Fee         string    `json:"fee,omitempty"`
	Memo        string    `json:"memo,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	TxMsgs      []TxMsg   `json:"tx_msgs"`
}
type TxMsg struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
	//Success    bool   `json:"success"`
	Log        string `json:"log"`
	From       string `json:"from"`
	To         string `json:"to"`
	Amount     string `json:"amount"`
	UnlockTime string `json:"unlock_time"`
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
	ChainID            string    `json:"chain_id"`
	Height             string    `json:"height"`
	Time               time.Time `json:"time"`
	LastBlockID        BlockID   `json:"last_block_id"`
	LastCommitHash     string    `json:"last_commit_hash"`     // commit from validators from the last block
	DataHash           string    `json:"data_hash"`            // transactions
	ValidatorsHash     string    `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash string    `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      string    `json:"consensus_hash"`       // consensus params for current block
	AppHash            string    `json:"app_hash"`             // state after txs from the previous block
	LastResultsHash    string    `json:"last_results_hash"`    // root hash of all results from the txs from the previous block
	EvidenceHash       string    `json:"evidence_hash"`        // evidence included in the block
	ProposerAddress    string    `json:"proposer_address"`     // original proposer of the block

	// basic block info
	//ChainID  string    `json:"chain_id"`
	//Height   string    `json:"height"`
	//Time     time.Time `json:"time"`
	//NumTxs   string    `json:"num_txs"`
	//TotalTxs string    `json:"total_txs"`
	//// prev block info
	//LastBlockID BlockID `json:"last_block_id"`
	//// hashes of block data
	//LastCommitHash string `json:"last_commit_hash"` // commit from validators from the last block
	//DataHash       string `json:"data_hash"`        // transactions
	//// hashes from the app output from the prev block
	//ValidatorsHash     string `json:"validators_hash"`      // validators for the current block
	//NextValidatorsHash string `json:"next_validators_hash"` // validators for the next block
	//ConsensusHash      string `json:"consensus_hash"`       // consensus params for current block
	//AppHash            string `json:"app_hash"`             // state after txs from the previous block
	//LastResultsHash    string `json:"last_results_hash"`    // root hash of all results from the txs from the previous block
	//// consensus info
	//EvidenceHash    string `json:"evidence_hash"`    // evidence included in the block
	//ProposerAddress string `json:"proposer_address"` // original proposer of the block
}

// Single block
type ResponseBlock struct {
	BlockId struct {
		Hash string `json:"hash"`
	} `json:"block_id"`
	Block struct {
		Header ResponseBlockHeader `json:"header"` // The block's Header
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
}
type ResponseInfo struct {
	Response struct {
		Data             string `json:"data"`                //"data": "kava",
		LastBlockHeight  string `json:"last_block_height"`   //"last_block_height": "606142",
		LastBlockAppHash string `json:"last_block_app_hash"` //"last_block_app_hash": "1WjEIIx06SVdpiiJURl0o+hbFKWtsMA7Khd1K61BA8U="
	} `json:"response"`
}

func (proxy *ResponseBlock) toBlock() *Block {
	block := Block{
		Hash:       proxy.BlockId.Hash,
		ParentHash: proxy.Block.Header.LastBlockID.Hash,
		ChainID:    proxy.Block.Header.ChainID,
		Timestamp:  proxy.Block.Header.Time,
	}
	block.Height, _ = utils.ParseInt64(proxy.Block.Header.Height)

	for _, tx := range proxy.Block.Data.Txs {
		block.Transactions = append(block.Transactions, fmt.Sprintf("%X", tx.Hash()))
	}
	return &block
}

/*
{
    "hash": "184639D663EF8D061E9F4BBF56ECCC2AFDFDEDC34936B581573923FF4EFC75E1",
    "height": "1376105",
    "index": 0,
    "tx_result": {
      "code": 0,
      "data": null,
      "log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"delegate\",\"attributes\":[{\"key\":\"validator\",\"value\":\"kavavaloper1dede4flaq24j2g9u8f83vkqrqxe6cwzrxt5zsu\"},{\"key\":\"amount\",\"value\":\"763497000\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"delegate\"},{\"key\":\"module\",\"value\":\"staking\"},{\"key\":\"sender\",\"value\":\"kava19fng49x6t0l87kq63e4xd7q2qjw85rd780rgvc\"}]}]}]",
      "info": "",
      "gasWanted": "200000",
      "gasUsed": "117033",
      "events": [
        {
          "type": "message",
          "attributes": [
            {
              "key": "YWN0aW9u",
              "value": "ZGVsZWdhdGU="
            }
          ]
        },
        {
          "type": "delegate",
          "attributes": [
            {
              "key": "dmFsaWRhdG9y",
              "value": "a2F2YXZhbG9wZXIxZGVkZTRmbGFxMjRqMmc5dThmODN2a3FycXhlNmN3enJ4dDV6c3U="
            },
            {
              "key": "YW1vdW50",
              "value": "NzYzNDk3MDAw"
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            {
              "key": "bW9kdWxl",
              "value": "c3Rha2luZw=="
            },
            {
              "key": "c2VuZGVy",
              "value": "a2F2YTE5Zm5nNDl4NnQwbDg3a3E2M2U0eGQ3cTJxanc4NXJkNzgwcmd2Yw=="
            }
          ]
        }
      ],
      "codespace": ""
    }*/
type ResponseTx struct {
	Height   string `json:"height"`
	TxHash   string `json:"hash"`
	Index    int    `json:"index"`
	TxResult struct {
		Code      int    `json:"code"` //code为0才是正确结果，并且log无法解析为对应结构，错误前提下会提示错误错误详情
		Log       string `json:"log"`
		Info      string `json:"info"`
		GasWanted string `json:"gasWanted"`
		GasUsed   string `json:"gasUsed"`
	} `json:"tx_result"`
	Tx        string    `json:"tx"`
	Timestamp time.Time `json:"_"`
}
type TxLog struct {
	MsgIndex int `json:"msg_index"`
	//Success  bool      `json:"success"` //缺少这个结构
	Log    string    `json:"log"`
	Events []TxEvent `json:"events"`
}
type TxEvent struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}

func (o *TxLog) getSender() string {
	for _, event := range o.Events {
		if event.Type == "message" {
			for _, attr := range event.Attributes {
				if attr.Key == "sender" {
					return attr.Value
				}
			}
		}
	}
	return ""
}
func (o *TxLog) getTxType() string {
	for _, event := range o.Events {
		if event.Type == "message" {
			for _, attr := range event.Attributes {
				if attr.Key == "module" {
					return attr.Value
				}
			}
		}
	}
	return ""
}
func (o *TxLog) getAction() string {
	for _, event := range o.Events {
		if event.Type == "message" {
			for _, attr := range event.Attributes {
				if attr.Key == "action" {
					return attr.Value
				}
			}
		}
	}
	return ""
}
func (o *TxLog) getReceiver(action string) string {
	for _, event := range o.Events {
		if action == TypeMsgSend && event.Type == "transfer" {
			for _, attr := range event.Attributes {
				if attr.Key == "recipient" {
					return attr.Value
				}
			}
		}
		if action == TypeMsgDelegate && event.Type == "validator" {
			for _, attr := range event.Attributes {
				if attr.Key == "recipient" {
					return attr.Value
				}
			}
		}
	}
	return ""
}
func (o *TxLog) getAmount() string {
	for _, event := range o.Events {
		if event.Type == "transfer" {
			for _, attr := range event.Attributes {
				if attr.Key == "amount" {
					return attr.Value
				}
			}
		}
	}
	return ""
}
func (proxy *ResponseTx) toTransaction(cdc *codec.Codec) (*Transaction, error) {
	tx := &Transaction{
		Hash:    proxy.TxHash,
		RawLogs: proxy.TxResult.Log,
	}
	tx.Timestamp = proxy.Timestamp
	tx.GasWanted, _ = utils.ParseInt64(proxy.TxResult.GasWanted)
	tx.GasUsed, _ = utils.ParseInt64(proxy.TxResult.GasUsed)
	tx.BlockHeight, _ = utils.ParseInt64(proxy.Height)
	if proxy.TxResult.Log == "" {
		return nil, fmt.Errorf("log nil")
	}
	if proxy.TxResult.Code != 0 {
		return nil, fmt.Errorf("虚假充值:%s,log:%s", proxy.TxHash, proxy.TxResult.Log)
	}
	//先解析交易，再解析log
	if proxy.Tx != "" {
		//	proxy.Tx = "zQEoKBapCkGoo2GaChT3lS3R+QohwgAGo1zSA0IN0TJ6VhIUMZoolMlyAnVbJqaXBc5BVAcRwyEaDwoFdWthdmESBjMyMTQ1NhISCgwKBXVrYXZhEgM1MDAQwJoMGmoKJuta6YchA2d7hwZhgksv6M7WOPU+yt4X4xNGAWhpHjEWJa9bVtZgEkDNC3+MjpDJTPa5+ItPgCTaC6t7/RnUSTmfVRtKUSoZv0V9VoG6bpE6eKgyt6mLziCh9sAD6qmakAW1Cwqi1tXKIgRydXN0"
		if raw, err := utils.Base64Decode([]byte(proxy.Tx)); err == nil {
			stdTx := auth.StdTx{}
			if err := cdc.UnmarshalBinaryLengthPrefixed(raw, &stdTx); err == nil {
				tx.Fee = stdTx.Fee.Amount.String()
				tx.Memo = stdTx.Memo
				tx.Type = "auth/StdTx"
				amt := stdTx.Fee.GasPrices().AmountOf(MainnetDenom)
				tx.GasPrice = amt.Uint64()
			} else {
				fmt.Printf("%v", err)
			}
		}
	}
	txLogs := make([]TxLog, 0)
	//解析log
	err := json.Unmarshal([]byte(proxy.TxResult.Log), &txLogs)
	if err != nil {
		return nil, err
	}
	if len(txLogs) == 0 {
		return nil, fmt.Errorf("don't hava any log")
	}
	for _, txLog := range txLogs {
		switch txLog.getTxType() {
		case "bank":
			txmsg := TxMsg{
				Index: txLog.MsgIndex,
				//Success: txLog.Success,
			}
			action := txLog.getAction()
			txmsg.Type = txLog.getTxType()
			txmsg.From = txLog.getSender()
			txmsg.To = txLog.getReceiver(action)
			txmsg.Amount = txLog.getAmount()
			tx.TxMsgs = append(tx.TxMsgs, txmsg)
		}
	}
	return tx, nil
}
