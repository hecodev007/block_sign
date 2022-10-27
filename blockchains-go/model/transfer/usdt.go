package transfer

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

type UsdtUnspents struct {
	Code    int        `json:"code"`
	Data    []UsdtUtxo `json:"data"`
	Message string     `json:"message"`
}

type UsdtUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
}

type UsdtOrderRequest struct {
	OrderRequestHead
	FromAddress  string                  `json:"fromAddress"`
	ToAddress    string                  `json:"toAddress"`
	Amount       int64                   `json:"amount,omitempty"`
	Fee          int64                   `json:"fee,omitempty"`
	OrderAddress []*UsdtOrderAddrRequest `json:"order_address,omitempty"`
}

type UsdtOrderCollectRequest struct {
	OrderRequestHead
	FromAddress  string                  `json:"from_address"`
	ToAddress    string                  `json:"to_address"`
	Amount       int64                   `json:"amount,omitempty"`
	Fee          int64                   `json:"fee,omitempty"`
	OrderAddress []*UsdtOrderAddrRequest `json:"order_address,omitempty"`
}

type UsdtOrderAddrRequest struct {
	Dir          DirType `json:"dir"`
	Address      string  `json:"address"`
	Amount       int64   `json:"amount"`
	TxID         string  `json:"txId"`
	Vout         int     `json:"vout"`
	TokenAmount  int64   `json:"tokenAmount,omitempty"`
	ScriptPubKey string  `json:"scriptPubKey"`
}

//====================手续费请求结果====================
type UsdtGasResult struct {
	FastestFee  int64 `json:"fastestFee"`
	HalfHourFee int64 `json:"halfHourFee"`
	HourFee     int64 `json:"hourFee"`
}

//====================手续费请求结果====================

//====================usdt验证地址====================
type UsdtAddressResp struct {
	Code    int                `json:"code"`
	Data    *UsdtAddressResult `json:"data"`
	Message string             `json:"message"`
}
type UsdtAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================usdt验证地址====================

//usdt余额信息

type UsdtBalanceResp struct {
	Code    int              `json:"code"`
	Data    *UsdtBalanceData `json:"data"`
	Message string           `json:"message"`
}

type UsdtBalanceData struct {
	BalanceFloat        decimal.Decimal `json:"balance"`        //浏览器余额
	RealBalanceFloat    decimal.Decimal `json:"realBalance"`    //真实金额
	PendingBalanceFloat decimal.Decimal `json:"pendingBalance"` //接收中金额
	LockFloat           decimal.Decimal `json:"lock"`           //发送中的金额
}

//usdt余额信息

func DecodeUsdtBalanceResp(ds []byte) (*UsdtBalanceResp, error) {
	ri := &UsdtBalanceResp{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeUsdtGasResult(ds []byte) (*UsdtGasResult, error) {
	ri := &UsdtGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeUsdtAddressResult(data []byte) *UsdtAddressResp {
	if len(data) != 0 {
		result := new(UsdtAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
