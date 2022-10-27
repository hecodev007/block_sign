package transfer

type CoinCollectToken struct {
	Coin  string   `json:"coin"`  // 币种名字
	MchId int64    `json:"mchId"` // 商户Id
	From  []string `json:"from"`
	To    string   `json:"to"` //指定冷地址
}

type CoinTransferFee struct {
	Coin     string `json:"coin"`
	To       string `json:"to"`
	MchId    int64  `json:"mchId"`
	FeeFloat string `json:"feeFloat"`
}

/*
查找代币地址是否有足够的手续费
*/
type CoinFindAddressFee struct {
	Coin    string `json:"coin"`
	MchId   int64  `json:"mchId"`
	Address string `json:"address"`
}
