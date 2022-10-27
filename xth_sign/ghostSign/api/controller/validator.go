package controller

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderNo  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//固定的几个参数
type Header struct {
	OrderNo  string `json:"order_no" binding:"required"`
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
	TxIns        []UcaTxInTpl  `json:"txIns" binding:"required"` //如果是
	TxOuts       []UcaTxOutTpl `json:"txOuts" binding:"required"`
	ChangeAddr   string        `json:"changeAddr"` //找零地址
	Fee          int64         `json:"fee" binding:"required,min=1000,max=10000000"`
	ExpiryHeight uint32        `json:"expiryHeight"`
}

//utxo模板
type UcaTxInTpl struct {
	FromAddr string `json:"fromAddr" binding:"required"` //来源地址
	//FromPrivkey      string `json:"fromPrivkey"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid   string `json:"fromTxid" binding:"required"`   //来源UTXO的txid
	FromIndex  uint32 `json:"fromIndex"`                     //来源UTXO的txid 地址的下标
	FromAmount int64  `json:"fromAmount" binding:"required"` //来源UTXO的txid 对应的金额
	//FromRedeemScript string `json:"fromRedeemScript"` //多签脚本
}

//输出模板
type UcaTxOutTpl struct {
	ToAddr   string `json:"toAddr" binding:"required"`   //txout地址
	ToAmount int64  `json:"toAmount" binding:"required"` //txout金额
}

type SignReturns struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Header
	Data   interface{} `json:"data"`
	TxHash string      `json:"txid"`
}

//zcash end

//telos
type TelosSignParams struct {
	Header
	Data *TelosSignParams_Data `json:"data"`
}
type TelosSignReturns struct {
	Header
	Data   interface{} `json:"data"`
	TxHash string      `json:"hash"`
}
type TelosSignParams_Data struct {
	Id          int64  `json:"id,omitempty"` //可以没有 暂时没用到
	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Token       string `json:"token" binding:"required"`    //telos主币是：“eosio.token”
	Quantity    string `json:"quantity" binding:"required"` //“1.001 TLOS”
	Memo        string `json:"memo,omitempty" binding:"required"`
	SignPubKey  string `json:"sign_pubkey" binding:"required"`
	BlockID     string `json:"block_id" binding:"required"` //最新10w个高度内的一个block ID,like:“0637f2d29169db2dfd3dfee61982edee74fa193bb8648b6419ed2749b08ed7d6”(所属高度104329938)
}
