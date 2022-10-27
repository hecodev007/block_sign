package cfx

import "github.com/shopspring/decimal"

type TxScan struct {
	Code int `json:"code"` //0成功，>0失败
	Message string `json:"message"`
	Hash string `json:"hash"`
	Nonce string `json:"nonce"`
	Value string `json:"value"`
	GasPrice string `json:"gasPrice"`
	Status int `json:"status"`  //0成功，1失败
	EpochHeight int64 `json:"epochHeight"`
	GasCoveredBySponsor bool `json:"gasCoveredBySponsor"`
	StorageCoveredBySponsor bool `json:"storageCoveredBySponsor"`
	GasFee decimal.Decimal `json:"gasFee"`
}
