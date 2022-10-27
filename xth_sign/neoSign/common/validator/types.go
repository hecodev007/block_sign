package validator

import "github.com/shopspring/decimal"

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//固定的几个参数
type Header struct {
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

type SignParams struct {
	Header
	SignParams_data
}

type SignParams_data struct {
	Type string `json:"type"` //默认contract,指定claim
	TxIns  []DogeTxInTpl  `json:"tx_ins" binding:"required"` //如果是
	TxOuts []DogeTxOutTpl `json:"tx_outs" binding:"required"`
}
type DogeTxInTpl struct {
	FromAddr         string `json:"from_addr"`                   //来源地址
	FromPrivkey      string `json:"from_privkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"from_txid"`                   //来源UTXO的txid
	FromIndex        int    `json:"from_index"`                  //来源UTXO的txid 地址的下标
	FromAmountInt64  int64  `json:"from_amount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"from_redeemScript,omitempty"` //多签脚本
	Assert string `json:"assert"`
}

//输出模板
type DogeTxOutTpl struct {
	ToAddr        string `json:"to_addr"`   //txout地址
	ToAmountInt64 int64  `json:"to_amount"` //txout金额
	Assert string `json:"assert"`
}
type SignReturns struct {
	Header
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TxHash  string      `json:"txid"`
}

type GetUtxos struct {
	Addr     string `json:"addr"  binding:"required"`
	Num      int    `json:"num" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}
type GetUtxosReturn struct {
	Code     int64      `json:"code"`
	Message  string     `json:"message"`
	Address  string     `json:"address"`
	Balances []*Balance `json:"balance"`
}

type Balance struct {
	Asset_symbol string          `json:"asset_symbol"`
	Asset_hash   string          `json:"asset_hash"`
	Asset        string          `json:"asset"`
	Amount       decimal.Decimal `json:"amount"`
	Unspent      []*Unspent      `json:"unspent"`
}
type Unspent struct {
	Value decimal.Decimal `json:"value"`
	Txid  string          `json:"txid"`
	N     int64           `json:"n"`
}
