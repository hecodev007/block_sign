package validator

import (
	"github.com/eoscanada/eos-go"
)

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
	CoinName string `json:"coin_name" `
}

type SignParams struct {
	SignHeader
	SignParams_Data
}

type SignParams_Data struct {
	Ins []*TxInTpl  `json:"txIns"`
	Outs []*TxOutTpl `json:"txOuts"`
}
//输入模板
type TxInTpl struct {
	FromAddr         string `json:"fromAddr"  binding:"required"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"  binding:"required"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmountInt64  int64  `json:"fromAmount"  binding:"required"`                 //来源UTXO的txid 对应的金额
	//暂不支持FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}
//输出模板
type TxOutTpl struct {
	ToAddr   string `json:"toAddr"  binding:"required"`   //txout地址
	ToAmountInt64 int64  `json:"toAmount"  binding:"required"` //txout金额
}

type SignReturns_data struct {
	eos.PackedTransaction
	TxHash interface{} `json:"txid"`
}

type SignReturns struct {
	SignHeader
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    string `json:"data"`
	Txid string `json:"txid"`
}

type TransferReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"txid"` //txid
}

type GetBalanceParams struct {
	CoinName string `json:"coin_name"`
	Token string `json:"token"`
	Address  string `json:"address" binding:"required"`
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
