package transfer

import (
	"encoding/json"
	"fmt"
)

type CeloOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	IsCollect   int    `json:"is_collect"` // 0: 表示正常交易 1： 表示归集
	// GasLimit 	uint64	`json:"gas_limit"`
	// Data 		string	`json:"data"`
}

func DecodeCeloTransferResp(data []byte) map[string]interface{} {
	var result map[string]interface{}
	if len(data) != 0 {
		err := json.Unmarshal(data, &result)
		if err == nil {
			return result
		} else {
			fmt.Printf("parse response data error,err=%v", err)
		}
	}
	return nil
}
