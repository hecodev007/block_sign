package validator

import (
	"encoding/json"

	algomodel "github.com/algorand/go-algorand-sdk/client/algod/models"

	"github.com/shopspring/decimal"
)

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}
type SignHeader struct {
	//MchId    int64  `json:"mch_id" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//////

//telos
type TelosSignParams struct {
	SignHeader
	*TelosSignParams_Data
	TransactionParams algomodel.TransactionParams `json:"transaction_params"`
}

func (p *TelosSignParams) String() string {
	str, _ := json.Marshal(p)
	return string(str)
}

type TelosSignParams_Data struct {
	FromAddress string          `json:"from_address" binding:"required"`
	ToAddress   string          `json:"to_address" binding:"required"`
	Value       decimal.Decimal `json:"value" binding:"required"`
	Fee         decimal.Decimal `json:"fee"`
	Assert      decimal.Decimal `json:"assert"`
	//Timestamp   int64           `json:"timestamp"` //

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
	Data    string `json:"data"` //txid
}

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
