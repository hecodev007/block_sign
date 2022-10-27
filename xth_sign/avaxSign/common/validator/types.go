package validator

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderNo  string `json:"orderNo" binding:"required"`
	MchName  string `json:"mchName" binding:"required"`
	CoinName string `json:"coinName" binding:"required"`
}

//固定的几个参数
type Header struct {
	OrderNo  string `json:"orderNo" binding:"required"`
	MchName  string `json:"mchName" binding:"required"`
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
	Utxos []string `json:"utxos" binding:"required"` //如果是
	//FromAddr   string   `json:"fromAddr" ` //从utxo里面提取address
	ToAddr     string `json:"toAddr" binding:"required"`
	ChangeAddr string `json:"changeAddr"`
	Amount     uint64 `json:"amount" binding:"required,min=1000"`
	Fee        uint64 `json:"fee" binding:"required,min=100000,max=5000000000"`
}

type SignReturns struct {
	Header
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TxHash  string      `json:"txid"`
}
