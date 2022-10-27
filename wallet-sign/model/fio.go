package model

type FioTransferParams struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//Memo 		string `json:"memo"`
}

type FioSignParams struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	HeadBlockId string `json:"head_block_id"`
	ChainId     string `json:"chain_id"`
}
