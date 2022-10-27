package controller

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

type ZcashSignParams struct {
	SignHeader
	Data interface{} `json:"data"`
}

type ZcashSignReturns struct {
	SignHeader
	Data   interface{} `json:"data"`
	TxHash string      `json:"hash"`
}

//zcash end

//telos
type TelosSignParams struct {
	SignHeader
	Data *TelosSignParams_Data `json:"data"`
}
type TelosSignReturns struct {
	SignHeader
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
