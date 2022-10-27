package transfer

type RubOrderReq struct {
	ChangeAddress string                       `json:"changeAddress"` //找零地址
	Fee           int64                        `json:"fee,omitempty"` //手续费
	Tos           []RubyChainParamsTransferTos `json:"tos"`
	ApplyId       int64                        `json:"applyId"`      //商户ID
	OuterOrderNo  string                       `json:"outerOrderNo"` //外部订单号
	OrderNo       string                       `json:"orderNo"`      //内部订单号
	MchName       string                       `json:"mchName"`      //商户名称
	CoinName      string                       `json:"coinName"`     //币种名称

}

type RubyChainParamsTransferTos struct {
	ToAddress string `json:"toAddress"` //接收地址
	ToAmount  int64  `json:"toAmount"`  //接收金额
}
