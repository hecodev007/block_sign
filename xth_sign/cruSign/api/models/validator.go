package models

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderNo  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//固定的几个参数
type Header struct {
	OrderNo  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

type CreateAddressReturns struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    CreateAddressReturns_data `json:"data"`
}

type CreateAddressReturns_data struct {
	CreateAddressParams
	Address []string `json:"address"`
}

type SignParams struct {
	Header
	SignParams_data
}

type SignParams_data struct {
	From               string `json:"fromAddr"` //from
	To                 string `json:"toAddr"`
	Amount             uint64 `json:"amount"`
	Nonce              uint64 `json:"nonce"`
	Fee                uint64 `json:"fee"`
	SpecVersion        uint32 `json:"specVersion"`
	TransactionVersion uint32 `json:"transactionVersion"`
	GenesisHash        string `json:"genesisHash"`
	BlockHash          string `json:"blockHash"`
	BlockNumber        uint64 `json:"blockNumber"`
	CallId             string `json:"callId"`
}

type SignReturns struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Header
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}
