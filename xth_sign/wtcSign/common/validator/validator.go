package validator

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int64  `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

type CreateAddressReturns struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"` //数值
}

type SignHeader struct {
	//MchId    int64  `json:"mch_id" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//////

//zcash
type ZcashCreateAddressReturns struct {
	Code    int                            `json:"code"`
	Message string                         `json:"message"`
	Data    ZcashCreateAddressReturns_data `json:"data"`
}

type ZcashCreateAddressReturns_data struct {
	CreateAddressParams
	Address []string `json:"address"`
}

type ZcashSignParams struct {
	SignHeader
	SignParams_Data
}

type ZcashSignReturns struct {
	SignHeader
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}
type SignParams_Data struct {
	Nonce       uint64           `json:"nonce"`
	FromAddress string           `json:"from_address" binding:"required"`
	ToAddress   string           `json:"to_address" binding:"required"`
	Token       string           `json:"contract_address"`
	TokenName   string           `json:"token"`
	Value       decimal.Decimal  `json:"value" binding:"required"` //
	GasLimit    uint64           `json:"gas_limit"`
	GasPrice    *decimal.Decimal `json:"gas_price"`
}

//zcash end

//telos
type TelosSignParams struct {
	SignHeader
	SignParams_Data
}

func (p *TelosSignParams) String() string {
	str, _ := json.Marshal(p)
	return string(str)
}

type TelosSignReturns_data struct {
	Data   string      `json:"data"`
	TxHash interface{} `json:"txid"`
}

type TelosSignReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	TelosSignReturns_data
}

type TelosTransferReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Rawtx   string `json:"rawtx"`
	Txhash  string `json:"txid"` //txid
}

type ValidAddressParams struct {
	Address string `json:"address"`
}
type ValidAddressReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"` //数值
}

type GetBalanceParams struct {
	Token   string `json:"contract_address"`
	Address string `json:"address"`
}
type GetBalanceReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
