package validator

import "github.com/eoscanada/eos-go"

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"order_no" binding:"required"`
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

type SignHeader struct {
	MchId    string `json:"mch_no" `
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" `
	CoinName string `json:"coin_name" binding:"required"`
}

type SignParams struct {
	SignHeader
	SignParams_Data
}
type SignParams_Data struct {
	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Token       string `json:"token"`                       //telos主币是：“eosio.token”
	Quantity    string `json:"quantity" binding:"required"` //“1.001 TLOS”
	Memo        string `json:"memo,omitempty"`
	SignPubKey  string `json:"sign_pubkey" binding:"required"`
	BlockID     string `json:"block_id" ` //最新10w个高度内的一个block ID,like:“0637f2d29169db2dfd3dfee61982edee74fa193bb8648b6419ed2749b08ed7d6”(所属高度104329938)
}

type SignReturns_data struct {
	eos.PackedTransaction
	TxHash interface{} `json:"txid"`
}

type SignReturns struct {
	SignHeader
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    SignReturns_data `json:"data"`
}

type TransferReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"txid"` //txid
}

type GetBalanceParams struct {
	CoinName string `json:"coin_name"`
	Address  string `json:"address" binding:"required"`
	Token    string `json:"contract_address" binding:"required"`
	Params   Params `json:"params"`
}
type Params struct {
	Symbol string `json:"symbol"` //币缩写 eg:bos
}
type GetBalanceReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"` //数值
}
type ValidAddressParams struct {
	Address string `json:"address"`
}
type ValidAddressReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"` //数值
}
