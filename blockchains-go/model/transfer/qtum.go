package transfer

import "encoding/json"

type QtumOrderRequest struct {
	OrderRequestHead
	FromAddress   string                `json:"from_address"`
	ChangeAddress string                `json:"change_address"`
	OrderAddress  []QtumOrderAddressReq `json:"order_address"`
	Token         string                //为空表示qtum转账
}

type QtumOrderAddressReq struct {
	Address  string `json:"address"`
	Amount   int64  `json:"amount"`
	Quantity string `json:"quantity"`
}

func DecodeQtumAddressResult(data []byte) *QtumAddressResp {
	if len(data) != 0 {
		result := new(QtumAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

type QtumAddressResp struct {
	Code    int               `json:"code"`
	Data    *BsvAddressResult `json:"data"`
	Message string            `json:"message"`
}
