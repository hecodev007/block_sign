package model

type CoinBalance struct {
	CoinName        string `json:"coinName"`
	Balance         string `json:"balance"`
	TokenName       string `json:"tokenName"`
	ContractAddress string `json:"contractAddress"`
	FeeAddress      string `json:"feeAddress"`
	ActivityBalance string `json:"activityBalance"`         //冷地址可用余额
	LiquidBalance   string `json:"liquidBalance,omitempty"` //冷地址可用余额
	/*
		func: 说是要获取前20个地址的总余额作为可余额
		date: 2021-03-15 write by jun
		author: flynn
	*/
	TopsTwentyBalance string `json:"topsTwentyBalance"`
}

type CoinAddrMaxBalance struct {
	CoinName  string `json:"coinName"`
	TokenName string `json:"tokenName"`
	Balance   string `json:"balance"`
	Address   string `json:"address"`
}
