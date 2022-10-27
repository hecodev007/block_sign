package transfer

import (
	"encoding/json"
)

// gxc订单请求
type GxcOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"` //发送地址
	ToAddress   string `json:"to_address"`   //接收地址
	ToAmount    int64  `json:"amount"`       //接收金额
	Memo        string `json:"memo"`         //memo
	Fee         int64  `json:"fee"`          //手续费
	PublicKey   string `json:"publicKey"`    //公钥
}

type GxcOrderRequestHot struct {
	OrderRequestHead
	FromAccount string `json:"fromAccount"`   //发送地址
	ToAccount   string `json:"toAccount"`     //接收地址
	Memo        string `json:"memo"`          //memo
	Amount      string `json:"amount"`        //接收金额，浮点字符
	Fee         string `json:"fee,omitempty"` //手续费
	PublicKey   string `json:"publicKey"`     //公钥

}

type GxcAddressResp struct {
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

type GxcTransferResp struct {
	Code    int                `json:"code"`
	Data    *GxcTransferResult `json:"data"`
	Message string             `json:"message"`
}
type GxcTransferResult struct {
	Txid string `json:"txid"`
	Fee  string `json:"fee"`
	Memo string `json:"memo"`
}

func DecodeGxcTransferResp(data []byte) *GxcTransferResp {
	result := &GxcTransferResp{
		Code:    -1,
		Data:    nil,
		Message: "",
	}
	if len(data) != 0 {
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

func DecodeGxcAddressResp(data []byte) (*GxcAddressResp, error) {
	result := &GxcAddressResp{
		Code:    -1,
		Data:    "false",
		Message: "",
	}
	err := json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
