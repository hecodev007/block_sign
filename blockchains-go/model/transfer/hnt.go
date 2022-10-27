package transfer

import (
	"encoding/json"
	"fmt"
)

type HntOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	Nonce       int64  `json:"nonce"` //热钱包出账不用设置，冷签需要设置
}

type HntTransferHotResp struct {
	Code int `json:"code"`
	Data struct {
		Txid     string `json:"txid"`
		FeeFloat string `json:"fee_float"`
	} `json:"data"`
	Message string `json:"message"`
}

func DecodeHntTransferHotResp(data []byte) (*HntTransferHotResp, error) {
	var thr HntTransferHotResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return nil, fmt.Errorf("parse transfer hot resp data error,Err=%v", err)
	}
	return &thr, nil
}
