package transfer

import "encoding/json"

type CkbOrderRequest struct {
	OrderRequestHead
	FeeString    string                   `json:"feestr"`
	IsForce      bool                     `json:"is_force"`
	OrderAddress []map[string]interface{} `json:"order_address"`
}

//type CkbOrderAddress struct {
//	Dir     int    `json:"dir"` // 0-> from, 1-> to, 2-> change
//	Address string `json:"address"`
//	Quantity	string `json:"quantity"`
//}

type CkbAddressResp struct {
	Code    int               `json:"code"`
	Data    *CkbAddressResult `json:"data"`
	Message string            `json:"message"`
}
type CkbAddressResult struct {
	Vaild bool `json:"vaild"`
}

func DecodeCkbAddressResult(data []byte) *CkbAddressResp {
	if len(data) != 0 {
		result := new(CkbAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
