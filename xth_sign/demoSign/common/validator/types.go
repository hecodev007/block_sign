package validator

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderNo  string `json:"orderId" binding:"required"`
	MchName  string `json:"mchId" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
}

//固定的几个参数
type Header struct {
	OrderNo  string `json:"orderId" binding:"required"`
	MchName  string `json:"mchId" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
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
	TxIns  []DogeTxInTpl  `json:"txIns" binding:"required"` //如果是
	TxOuts []DogeTxOutTpl `json:"txOuts" binding:"required"`
}
type DogeTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmountInt64  int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}

//输出模板
type DogeTxOutTpl struct {
	ToAddr        string `json:"toAddr"`   //txout地址
	ToAmountInt64 int64  `json:"toAmount"` //txout金额
}
type SignReturns struct {
	Header
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TxHash  string      `json:"txid"`
}
