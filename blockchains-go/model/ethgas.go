package model

import "github.com/shopspring/decimal"

//{
//"status": "1",
//"message": "OK",
//"result": {
//"LastBlock": "11391719",
//"SafeGasPrice": "14",
//"ProposeGasPrice": "19",
//"FastGasPrice": "21"
//}
//}

type EthScanGasDataResult struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  *EthScanGasData `json:"result"`
}

type EthScanGasData struct {
	SafeGasPrice    decimal.Decimal `json:"SafeGasPrice"`
	ProposeGasPrice decimal.Decimal `json:"ProposeGasPrice"`
	FastGasPrice    decimal.Decimal `json:"FastGasPrice"`
}

type EthFeeResult struct {
	EthFee      string `json:"ethFee"`
	EthTokenFee string `json:"ethTokenFee"`
}
