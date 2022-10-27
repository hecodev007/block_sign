package util

import (
	"errors"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/common/cryp/crypto"
	"io"
	"math/big"
	"strings"
)

type GetAccountJson struct {
	Address string `json:"address"`
	ChainId string `json:"chainId"`
}

type GetBlockTxsJson struct {
	ChainId string `json:"chainId"`
	Height  string `json:"height"`
	Page    string `json:"page"`
	Size    string `json:"size"`
}

type Account struct {
	Addr            string   `json:"address"`
	Nonce           uint64   `json:"nonce"`
	Balance         *big.Int `json:"balance"`         // 系统基础货币 TUE，不为nil
	LocalCurrency   *big.Int `json:"localCurrency"`   // 本链第二货币（如果存在），可为nil
	StorageRoot     []byte   `json:"storageRoot"`     // 智能合约使用的存储，Trie(key: Hash, value: Hash)
	CodeHash        []byte   `json:"codeHash"`        // 合约代码的Hash
	LongStorageRoot []byte   `json:"longStorageRoot"` // 系统合约用来保存更灵活的数据结构, Trie(key: Hash, value: []byte)
}
type Transaction struct {
	ChainId      string `json:"chainId"`
	FromChainId  string `json:"fromChainId,omitempty"`
	ToChainId    string `json:"toChainId,omitempty"`
	From         string `json:"from"`
	To           string `json:"to"`
	Nonce        string `json:"nonce"`
	Value        string `json:"value"`
	Sig          string `json:"sig,omitempty"`
	Pub          string `json:"pub,omitempty"`
	Input        string `json:"input"`
	UseLocal     bool   `json:"useLocal"`
	Extra        string `json:"extra"` // 目前用来存交易类型，不存在时为普通交易，否则会对应特殊操作
	ExpireHeight int64  `json:"expireHeight,omitempty"`
}

func (tx Transaction) HashValue() ([]byte, error) {
	hasher := crypto.GetHash256()
	if _, err := tx.HashSerialize(hasher); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// 此处与rpcTx的Hash算法一致
func (tx Transaction) HashSerialize(w io.Writer) (int, error) {
	var toAddr string
	var fromAddr string
	if has0xPrefix(tx.To) {
		toAddr = tx.To[2:]
		toAddr = strings.ToLower(toAddr)
	} else {
		return 0, errors.New("hex string without 0x prefix")
	}

	if has0xPrefix(tx.From) {
		fromAddr = tx.From[2:]
		fromAddr = strings.ToLower(fromAddr)
	} else {
		return 0, errors.New("hex string without 0x prefix")
	}

	var input string
	if has0xPrefix(tx.Input) {
		input = tx.Input[2:]
		input = strings.ToLower(input)
	} else {
		return 0, errors.New("hex string without 0x prefix")
	}
	u := "0"
	if tx.UseLocal {
		u = "1"
	}

	extra := ""
	if len(extra) > 2 {
		extra = extra[2:]
	}

	str := []string{tx.ChainId, fromAddr, toAddr, tx.Nonce, u, tx.Value, input, extra}
	p := strings.Join(str, "-")
	return w.Write([]byte(p))
}
func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

type GetTxByHash struct {
	ChainId string `json:"chainId"`
	Hash    string `json:"hash"`
}

type GetBlockHeader struct {
	ChainId string `json:"chainId"`
	Height  string `json:"height"`
}

type PingJson struct {
	ChainId string `json:"chainId"`
}

type GetChainInfoJson struct {
	ChainId []int `json:"chainId"`
}

type GetStatsJson struct {
	ChainId string `json:"chainId"`
}

type GetTransactionsJson struct {
	ChainId     string `json:"chainId"`
	Address     string `json:"address"`
	StartHeight string `json:"startHeight"`
	EndHeight   string `json:"endHeight"`
}

type GetMultiStatsJson struct {
	ChainId string `json:"chainId"`
}

type GetCommitteeJson struct {
	ChainId string `json:"chainId"`
	Epoch   int    `json:"epoch"`
}
type CompileContractJson struct {
	ChainId  string `json:"chainId"`
	Contract string `json:"contract"`
}
