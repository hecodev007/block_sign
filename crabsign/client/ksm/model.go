package ksm

import (
	"github.com/shopspring/decimal"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
)

type Txparam struct {
	MchName            string          `json:"mch_name"`
	FromAddress        string          `json:"from_address"`
	ToAddress          string          `json:"to_address"`
	Amount             decimal.Decimal `json:"amount"`
	Nonce              uint64          `json:"nonce"`
	SpecVersion        uint32          `json:"spec_version"`
	TransactionVersion uint32          `json:"transaction_version"`
	GenesisHash        string          `json:"genesis_hash"`
	BlockHash          string          `json:"block_hash"`
	BlockNumber        uint64          `json:"block_number"`
	Meta               *types.Metadata
}

func SignTx(params *Txparam, privateKey string) (string, error) {

	extri, err := BuildTx(params, privateKey)
	if err != nil {
		return "", err
	}

	return types.EncodeToHexString(extri)
}
