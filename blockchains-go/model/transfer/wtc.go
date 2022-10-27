package transfer

import (
	"encoding/json"
	"fmt"
)

type WtcOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Value       string `json:"value"`
	//  token 转账 需要参数
	ContractAddress string `json:"contract_address"`
	Token           string `json:"token"` // 代币的名字，主链转账不传这个值
}

func DecodeWtcTransferResp(data []byte) map[string]interface{} {
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
