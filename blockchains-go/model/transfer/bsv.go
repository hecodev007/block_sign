package transfer

import "encoding/json"

type BsvOrderRequest struct {
	OrderRequestHead
	FromAddress  string                    `json:"from_address"`
	OrderAddress []*BsvOrderAddressRequest `json:"order_address"`
}

type BsvOrderAddressRequest struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

func DecodeBsvAddressResult(data []byte) *BsvAddressResp {
	if len(data) != 0 {
		result := new(BsvAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

type BsvAddressResp struct {
	Code    int               `json:"code"`
	Data    *BsvAddressResult `json:"data"`
	Message string            `json:"message"`
}

type BsvAddressResult struct {
	Isvalid bool `json:"isvalid"`
}
