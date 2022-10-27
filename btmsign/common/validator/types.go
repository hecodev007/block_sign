package validator

//固定的几个参数
type Header struct {
	OrderId  string `json:"order_no" binding:"required"`
	MchId    int    `json:"mch_id" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
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

type GetBalanceResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
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

type BtmSignParams struct {
	Header
	SignParams_databtm
}

type SignParams_databtm struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Fee         string `json:"fee"`
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

type ReqGetBalanceParams struct {
	CoinName        string      `json:"coin_name"`        // 币种主链的名字
	Address         string      `json:"address"`          //	需要获取余额的地址
	Token           string      `json:"token"`            // 	token的名字
	ContractAddress string      `json:"contract_address"` // 合约地址
	Params          interface{} `json:"params"`           // 特殊参数（如果有特殊参数，传入到这里面）
	MchId           string      `json:"mch_id"`
}

type ReqValidateAddrParams struct {
	Address string `json:"address"`
}
