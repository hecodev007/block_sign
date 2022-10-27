package model

import "encoding/json"

//商户返回调确认格式
type CallbackResp struct {
	Code int `json:"code"`
	//Message string `json:"message"`
}

func DecodeBCallbackResp(data []byte) *CallbackResp {
	result := &CallbackResp{
		Code: -1,
		//Message: "",
	}
	if len(data) != 0 {
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return result
}

type BaseCreateData struct {
	CoinName string `json:"coinName"`
	MchId    string `json:"mchId"`
	OrderId  string `json:"orderId"`
}
type CkbCreateData struct {
	BaseCreateData
	Inputs  []CkbInput  `json:"inputs"`
	Outputs []CkbOutput `json:"outputs"`
}
type CkbInput struct {
	Address string `json:"address"`
	Txid    string `json:"txid"`
	Index   int    `json:"index"`
	Amount  string `json:"amount"`
}

type CkbOutput struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

//       btm

type BtmCreateData struct {
	BaseCreateData
	Sources []BtmSource `json:"sources"`
}

type BtmSource struct {
	OutputId  string `json:"output_id"`
	SourceId  string `json:"source_id"`
	SourcePos int    `json:"source_pos"`
	Amount    int64  `json:"amount"`
}
