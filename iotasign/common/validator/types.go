package validator

//固定的几个参数
type Header struct {
	OrderId  string `json:"orderId" binding:"required"`
	MchId    string `json:"mchId" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
}

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num int `json:"num" binding:"min=1,max=50000"`
	Header
}

//创建账户参数
type CreateAddressReturns struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    CreateAddressReturns_data `json:"data"`
}

type CreateAddressReturns_data struct {
	CreateAddressParams
	Address []string `json:"address"`
}

//签名参数
type SignParams struct {
	Header
	SignParams_data
}

type SignParams_data struct {
	Ins  []*TxInTpl  `json:"txIns"`
	Outs []*TxOutTpl `json:"txOuts"`
}

//输入模板
type TxInTpl struct {
	FromAddr        string `json:"fromAddr"  binding:"required"`   //来源地址
	FromPrivkey     string `json:"fromPrivkey,omitempty"`          //来源地址地址对于的私钥，签名期间赋值
	FromTxid        string `json:"fromTxid"  binding:"required"`   //来源UTXO的txid
	FromIndex       uint32 `json:"fromIndex"`                      //来源UTXO的txid 地址的下标
	FromAmountInt64 int64  `json:"fromAmount"  binding:"required"` //来源UTXO的txid 对应的金额
	//暂不支持FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}

//输出模板
type TxOutTpl struct {
	ToAddr        string `json:"toAddr"  binding:"required"`   //txout地址
	ToAmountInt64 int64  `json:"toAmount"  binding:"required"` //txout金额
}

type SignReturns struct {
	Header
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TxHash  string      `json:"txid"`
}
