package reset

type ResetEthReq struct {
	OrderId string `json:"orderid"`
	Txid    string `json:"txid"`
}

type ResetEthGasReq struct {
	MinGasPrice int64 `json:"min_gas_price"` //0
	MaxGasPrice int64 `json:"max_gas_price"` //190000000000
	MinTokenGas int64 `json:"min_token_gas"` //75000
}

type ResetEthCollectReq struct {
	Coin string `json:"coin"` //币种
	Bof  string `json:"bof"`  //开始金额
}
