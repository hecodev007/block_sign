package transfer

type WbcOrderReq struct {
	OrderRequestHead
	ChangeAddress string                      `json:"changeAddress"` //找零地址
	Fee           int64                       `json:"fee,omitempty"` //手续费
	Tos           []WbcChainParamsTransferTos `json:"tos"`
}

type WbcChainParamsTransferTos struct {
	ToAddress string `json:"toAddress"` //接收地址
	ToAmount  int64  `json:"toAmount"`  //接收金额
}
