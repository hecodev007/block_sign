package iost

import (
	"encoding/json"
	"errors"
	"fmt"
	"iostsync/common/log"

	"github.com/shopspring/decimal"
)

func (rpc *RpcClient) BlockNumber() (int64, error) {
	chaininfo, err := rpc.GetChainInfo()
	if err != nil {
		return 0, err
	}
	return chaininfo.HeadBlock.IntPart(), nil
}

type ChainInfoResponse struct {
	HeadBlock decimal.Decimal `json:"head_block"`
}

func (rpc *RpcClient) GetChainInfo() (chainInfo *ChainInfoResponse, err error) {
	resp, err := rpc.Get(rpc.url + "/getChainInfo")
	if err != nil {
		return nil, err
	}
	chainInfo = new(ChainInfoResponse)
	err = json.Unmarshal(resp, chainInfo)
	return
}

type TransactionResponse struct {
	Status      string          `json:"status"`
	Transaction *Transaction    `json:"transaction"`
	BlockNumber decimal.Decimal `json:"block_number"`

	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (reponse *TransactionResponse) Error() error {
	if reponse.Code != 0 {
		return errors.New(reponse.Message)
	}
	return nil
}

type Transaction struct {
	Hash        string             `json:"hash"`
	Time        decimal.Decimal    `json:"time"`
	Expiration  decimal.Decimal    `json:"expiration"`
	GasRatio    int64              `json:"gas_ratio"`
	GasLimit    int64              `json:"gas_limit"`
	Delay       decimal.Decimal    `json:"delay"`
	ChainId     int64              `json:"chain_id"`
	Actions     []Action           `json:"actions"`
	Signers     []interface{}      `json:"signers"`
	Publisher   string             `json:"publisher"`
	ReferredTx  string             `json:"referred_tx"`
	AmountLimit []interface{}      `json:"amount_limit"`
	TxReceipt   TransactionReceipt `json:"tx_receipt"`
}
type Action struct {
	Contract   string `json:"contract"`
	ActionName string `json:"action_name"`
	Data       string `json:"data"`
}
type TransactionReceipt struct {
	TxHash     string      `json:"tx_hash"`
	GasUsage   int64       `json:"gas_usage"`
	RamUsage   interface{} `json:"ram_usage"`
	StatusCode string      `json:"status_code"` //SUCCESS
	Resturns   interface{} `json:"resturns"`
	Receipts   []struct {
		FuncName string          `json:"func_name"`
		Content  json.RawMessage `json:"content"`
	}
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (reponse *TransactionReceipt) Error() error {
	if reponse.Code != 0 {
		return errors.New(reponse.Message)
	}
	return nil
}
func (rpc *RpcClient) TransactionByHash(txhash string) (*TransactionResponse, error) {
	resp, err := rpc.Get(rpc.url + "/getTxByHash/" + txhash)
	if err != nil {
		return nil, err
	}
	result := new(TransactionResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, result.Error()
}
func (rpc *RpcClient) TransactionReceipt(txhash string) (*TransactionReceipt, error) {
	resp, err := rpc.Get(rpc.url + "/getTxReceiptByTxHash/" + txhash)
	if err != nil {
		return nil, err
	}
	result := new(TransactionReceipt)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}

	return result, result.Error()
}

type BlockReponse struct {
	Status  string `json:"status"`
	Block   *Block `json:"block"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (reponse *BlockReponse) Error() error {
	if reponse.Code != 0 {
		return errors.New(reponse.Message)
	}
	if reponse.Status != "IRREVERSIBLE" {
		return errors.New("pending block")
	}
	return nil
}

type Block struct {
	Hash                string          `json:"hash"`
	Version             string          `json:"version"`
	ParentHash          string          `json:"parent_hash"`
	TxMerkleHash        string          `json:"tx_merkle_hash"`
	TxReceiptMerkleMash string          `json:"tx_receipt_merkle_mash"`
	Number              decimal.Decimal `json:"number"`
	Witness             string          `json:"witness"`
	Time                decimal.Decimal `json:"time"`
	GasUsage            int64           `json:"gas_usage"`
	TxCount             decimal.Decimal `json:"tx_count"`
	Info                interface{}     `json:"info"`
	OrigInfo            string          `json:"orig_info"`
	Transactions        []*Transaction  `json:"transactions"`
}

func (rpc *RpcClient) BlockByNumber(h int64) (*Block, error) {
	resp, err := rpc.Get(fmt.Sprintf("%v/getBlockByNumber/%v/true", rpc.url, h))
	if err != nil {
		return nil, err
	}
	result := new(BlockReponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		log.Info(string(resp))
		return nil, err
	}
	if result.Error() != nil {
		//log.Info(string(resp))
	}
	return result.Block, result.Error()
}
