package luna

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tendermint/tendermint/types"

	"github.com/shopspring/decimal"
)

func (c *RpcClient) GetBlockCount() (int64, error) {
	resp, err := http.Get(fmt.Sprintf("%s/blocks/latest", c.url))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	blockResp := new(Block)
	err = json.Unmarshal(body, blockResp)
	if err != nil {
		return 0, err
	}
	if blockResp.Error != "" {
		return 0, errors.New(blockResp.Error)
	}
	return blockResp.Block.Header.Height.IntPart(), nil
}

type Block struct {
	BlockId BlockID    `json:"block_id"`
	Block   *BlockData `json:"block"`
	Error   string     `json:"error"`
}
type BlockData struct {
	Header ResponseBlockHeader `json:"header"`
	Data   struct {
		Txs types.Txs `json:"txs"`
	} `json:"data"`
}
type BlockID struct {
	Hash        string `json:"hash"`
	PartsHeader struct {
		Total int64  `json:"total"`
		Hash  string `json:"hash"`
	} `json:"parts"`
}
type ResponseBlockHeader struct {
	// basic block info
	ChainID string          `json:"chain_id"`
	Height  decimal.Decimal `json:"height"`
	Time    time.Time       `json:"time"`
	NumTxs  string          `json:"num_txs"`
	//TotalTxs string    `json:"total_txs"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`
}

func (c *RpcClient) GetBlockByHeight(height int64) (*Block, error) {
	resp, err := http.Get(fmt.Sprintf("%s/blocks/%v", c.url, height))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	blockResp := new(Block)
	err = json.Unmarshal(body, blockResp)
	if err != nil {
		return nil, err
	}
	if blockResp.Error != "" {
		return nil, errors.New(blockResp.Error)
	}
	return blockResp, nil
}

type Transaction struct {
	Hash        string    `json:"hash"`
	BlockHeight int64     `json:"block_height"`
	BlockHash   string    `json:"block_hash"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Value       int64     `json:"value"`
	Fee         int64     `json:"fee,omitempty"`
	Memo        string    `json:"memo,omitempty"`
	Success     bool      `json:"success"`
	Token       string    `json:"token"`
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	//UnlockTime string    `json:"unlock_time"`

	GasUsed   int64  `json:"gas_used,omitempty" `
	GasWanted int64  `json:"gas_wanted,omitempty"`
	GasPrice  string `json:"gas_price,omitempty"`
	RawLogs   string `json:"raw_logs,omitempty"`
	//TxMsgs    []TxMsg `json:"tx_msgs"`
}

type TxMsg struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Log     string `json:"log"`
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  string `json:"amount"`
	//UnlockTime string `json:"unlock_time"`
}

type TxResponse struct {
	Height decimal.Decimal `json:"height"`
	Txhash string          `json:"txhash"`
	Data   string          `json:"data"`
	RawLog string          `json:"raw_log"`
	Logs   []struct {
		Events []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"logs"`
	GasWanted string `json:"gas_wanted"`
	GasUsed   string `json:"gas_used"`
	Tx        struct {
		Type  string `json:"type"`
		Value struct {
			Msgs []struct {
				Type  string `json:"type"`
				Value struct {
					//MsgSend
					FromAddress string `json:"from_address"`
					ToAddress   string `json:"to_address"`
					Amounts     []struct {
						Denom  string          `json:"denom"`
						Amount decimal.Decimal `json:"amount"`
					} `json:"amount"`
					//MsgExecuteContract
					Sender     string `json:"sender"`
					Contract   string `json:"contract"`
					ExecuteMsg struct {
						Transfer struct {
							Recipient string          `json:"recipient"`
							Amount    decimal.Decimal `json:"amount"`
						} `json:"transfer"`
					} `json:"execute_msg"`
				} `json:"value"`
			} `json:"msg"`
			Fee struct {
				Amount []struct {
					Denom  string          `json:"denom"`
					Amount decimal.Decimal `json:"amount"`
				} `json:"amount"`
				Gas string `json:"gas"`
			} `json:"fee"`
			Signatures    []interface{} `json:"signatures"`
			Memo          string        `json:"memo"`
			TimeoutHeight string        `json:"timeout_height"`
		} `json:"value"`
	} `json:"tx"`
	Timestamp time.Time `json:"timestamp"`
	Events    []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Index bool   `json:"index"`
		} `json:"attributes"`
	} `json:"events"`
	Error string `json:"error"`
}

func (tx *TxResponse) ToTransactions() (txs []*Transaction, err error) {
	if tx.Logs == nil || len(tx.Logs) == 0 {
		return nil, errors.New(tx.RawLog)
	}
	for _, v := range tx.Events {
		for _, a := range v.Attributes {
			if a.Index == false {
				return nil, errors.New("tx.event.Attributes.index==faslse:" + tx.Txhash)
			}
		}
	}
	for _, msg := range tx.Tx.Value.Msgs {
		if msg.Type == "bank/MsgSend" {
			tmpTx := new(Transaction)
			tmpTx.Hash = strings.ToLower(tx.Txhash)
			tmpTx.Success = true
			tmpTx.BlockHeight = tx.Height.IntPart()
			tmpTx.From = msg.Value.FromAddress
			tmpTx.To = msg.Value.ToAddress
			if len(tx.Tx.Value.Fee.Amount) == 0 {
				tmpTx.Fee = 0
			} else {
				tmpTx.Fee = tx.Tx.Value.Fee.Amount[0].Amount.IntPart()
			}
			tmpTx.Timestamp = tx.Timestamp
			tmpTx.Memo = tx.Tx.Value.Memo
			//Fee := tx.Tx.Value.Fee.Amount[0].Amount.IntPart()
			for _, amount := range msg.Value.Amounts {
				transaction := new(Transaction)
				*transaction = *tmpTx
				transaction.Token = amount.Denom
				transaction.Value = amount.Amount.IntPart()
				txs = append(txs, transaction)
			}
		}
		if msg.Type == "wasm/MsgExecuteContract" {
			tmpTx := new(Transaction)
			tmpTx.Hash = strings.ToLower(tx.Txhash)
			tmpTx.Success = true
			tmpTx.BlockHeight = tx.Height.IntPart()
			tmpTx.From = msg.Value.Sender
			tmpTx.To = msg.Value.ExecuteMsg.Transfer.Recipient
			if len(tx.Tx.Value.Fee.Amount) == 0 {
				tmpTx.Fee = 0
			} else {
				tmpTx.Fee = tx.Tx.Value.Fee.Amount[0].Amount.IntPart()
			}
			tmpTx.Timestamp = tx.Timestamp
			tmpTx.Memo = tx.Tx.Value.Memo
			tmpTx.Token = msg.Value.Contract
			tmpTx.Value = msg.Value.ExecuteMsg.Transfer.Amount.IntPart()
			//Fee := tx.Tx.Value.Fee.Amount[0].Amount.IntPart()
			txs = append(txs, tmpTx)
		}
	}
	if len(txs) == 0 {
		return nil, errors.New("不支持的交易类型")
	}

	return txs, nil
}
func (c *RpcClient) GetRawTransaction(txid string) ([]*Transaction, error) {
	resp, err := http.Get(fmt.Sprintf("%s/txs/%v", c.url, txid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	txResponse := new(TxResponse)
	err = json.Unmarshal(body, txResponse)
	if err != nil {
		return nil, err
	}
	//log.Info(string(body))
	if len(txResponse.Logs) == 0 {
		return nil, errors.New(txResponse.RawLog)
	}
	//log.Info(xutils.String(txResponse))
	return txResponse.ToTransactions()
}

//
//func (c *RpcClient) GetTransaction(txid string) (*Result, error) {
//
//}
