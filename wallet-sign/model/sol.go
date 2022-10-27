package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

type SolTransferParams struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
}

type SolSignParams struct {
	ReqBaseParams
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	RecentBlockHash string `json:"recent_block_hash"`
}

type SolRecentBlockHash struct {
	Value SolRecentBHValue `json:"value"`
}
type SolRecentBHValue struct {
	RecentBlockHash string `json:"blockhash"`
}

func DecodeSolRecentBlockHash(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data length is 0")
	}
	var sbh SolRecentBlockHash
	err := json.Unmarshal(data, &sbh)
	if err != nil {
		return "", fmt.Errorf("json unmarshal recent block hash error,err=%v", err)
	}
	return sbh.Value.RecentBlockHash, nil
}
