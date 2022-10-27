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
	From       string `json:"from_addr"` //from
	To         string `json:"to_addr"`
	Amount     string `json:"amount"`
	Nonce      int64  `json:"nonce"`
	GasPremium int64  `json:"gas_premium"`
	GasFeeCap  int64  `json:"gas_fee_cap"`
	GasLimit   int64  `json:"gas_limit" `
}

type SignReturns struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Header
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}
