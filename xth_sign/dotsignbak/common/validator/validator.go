package validator

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
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
	Data interface{} `json:"data"`
}

type ZcashSignReturns struct {
	SignHeader
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}

//zcash end

//telos
type TelosSignParams struct {
	SignHeader
	*TelosSignParams_Data
}

func (p *TelosSignParams) String() string {
	str, _ := json.Marshal(p)
	return string(str)
}

type TelosSignParams_Data struct {
	FromAddress        string `json:"from_address"`
	ToAddress          string `json:"to_address"`
	Amount             decimal.Decimal `json:"amount"`
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
	GenesisHash        string `json:"genesis_hash"`
	BlockHash          string `json:"block_hash"`
	BlockNumber        uint64 `json:"block_number"`
	Meta *types.Metadata
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

type ValidAddressParams struct {
	Address string `json:"address"`
}
type ValidAddressReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"` //数值
}

type GetBalanceResponse struct{
	Code int `json:"code"`
	Message string `json:"message"`
	Data string `json:"data"`
}