package transfer

type VetOrderRequest struct {
	OrderRequestHead
	Data []VetData `json:"data"`
}

type VetData struct {
	SubData VetSubData `json:"data"`
}

type VetSubData struct {
	From            string      `json:"from"`
	Nonce           int64       `json:"nonce"`
	BlockNumber     int64       `json:"blockNumber"`
	ContractAddress string      `json:"contractAddress"`
	CoinName        string      `json:"coinName"`
	Tolist          []VetToList `json:"tolist"`
}

type VetToList struct {
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type VetRepay struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Repay  string `json:"repay"`
	Amount string `json:"amount"`
}
