package common

import "encoding/json"

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"orderId" binding:"required"`
	MchId    string `json:"mchId" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
}
type SignHeader struct {
	MchId    string `json:"mchId" binding:"required"`
	OrderId  string `json:"orderId" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
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

type ZenSignParams struct {
	SignHeader
	Data
}

func (z *ZenSignParams) String() string {
	str, _ := json.Marshal(z)
	return string(str)
}

type ZenSignReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	SignHeader
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}

func (z *ZenSignReturns) String() string {
	str, _ := json.Marshal(z)
	return string(str)
}

type Data struct {
	TxIns  []UcaTxInTpl  `json:"txIns" binding:"required"` //如果是
	TxOuts []UcaTxOutTpl `json:"txOuts" binding:"required"`
	//ChangeAddr   string        `json:"changeAddr"`
	//Fee          int64         `json:"fee" binding:"min=0,max=10000"`
	BlockHash   string `json:"block_hash"`
	BlockHeight int64  `json:"block_height"`
	//ExpiryHeight uint32        `json:"expiryHeight"`
}

//utxo模板
type UcaTxInTpl struct {
	FromAddr string `json:"fromAddr"` //来源地址
	//FromPrivkey      string `json:"fromPrivkey"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid   string `json:"fromTxid"`   //来源UTXO的txid
	FromIndex  uint32 `json:"fromIndex"`  //来源UTXO的txid 地址的下标
	FromAmount int64  `json:"fromAmount"` //来源UTXO的txid 对应的金额
	FromScript string `json:"from_script"`
	//FromRedeemScript string `json:"fromRedeemScript"` //多签脚本
}

//输出模板
type UcaTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
