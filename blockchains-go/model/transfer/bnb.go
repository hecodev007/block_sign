package transfer

import "encoding/json"

type BNBAddressResp struct {
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

type BNBOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"` //发送地址
	ToAddress   string `json:"to_address"`   //接收地址
	Quantity    string `json:"quantity"`     //金额，数量为 amount*10的八位
	Memo        string `json:"memo"`         //memo
	Token       string `json:"token"`        //token
	IsRetry     bool   `json:"is_retry"`     //是否重试，默认false，不是必填
}

type BNBBalanceReq struct {
	Address      string `json:"address"`
	ContractAddr string `json:"contract_addr,omitempty"`
	Decimal      int    `json:"decimal,omitempty"`
}

type BNBBalanceResp struct {
	Address  string           `json:"address"`
	Balances []*BalanceResult `json:"balances"`
}

type BalanceResult struct {
	Free   string `json:"free"`
	Frozen string `json:"frozen"`
	Locked string `json:"locked"`
	Symbol string `json:"symbol"`
}

type BNBCollectReq struct {
	OrderRequestHead
	FromAddrs    []string `json:"from_addrs"`
	ToAddr       string   `json:"to_addr"`
	ContractAddr string   `json:"contract_addr,omitempty"`
	Decimal      int      `json:"decimal"`
}

type BNBTransferFeeReq struct {
	OrderRequestHead
	FromAddr string   `json:"from_addr"`
	ToAddrs  []string `json:"to_addrs"`
	NeedFee  string   `json:"need_fee"`
}

func DecodeBNBAddressResp(data []byte) (*BNBAddressResp, error) {
	result := &BNBAddressResp{
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

func DecodeBNBBalanceResp(data []byte) (*BNBBalanceResp, error) {
	var result BNBBalanceResp
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
