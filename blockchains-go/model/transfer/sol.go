package transfer

import (
	"encoding/json"
	"fmt"
)

type SolOrderRequest struct {
	OrderRequestHead
	MchId                   string `json:"mchId"`
	OrderId                 string `json:"OrderId"`
	FromAddress             string `json:"from_address"`
	ToAddress               string `json:"to_address"` // '接收者地址'
	Amount                  string `json:"amount"`
	ContractAddress         string `json:"contract_address"` //主链币sol转账可以不传
	FeeAddress              string `json:"fee_address"`
	OriginalContractAddress string `json:"original_contract_address"`
}

func DecodeSolTransferResp(data []byte) map[string]interface{} {
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
