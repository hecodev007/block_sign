package bo

import "github.com/shopspring/decimal"

//eth API 合约交易
type EthAPIInternal struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  []InternalResult `json:"result"`
}

type InternalResult struct {
	BlockNumber     string          `json:"blockNumber"`
	TimeStamp       string          `json:"timeStamp"`
	From            string          `json:"from"`
	To              string          `json:"to"`
	Value           decimal.Decimal `json:"value"`
	ContractAddress string          `json:"contractAddress"`
	Input           string          `json:"input"`
	Type            string          `json:"type"` //只能认call，目前做强校验
	Gas             string          `json:"gas"`
	GasUsed         string          `json:"gasUsed"`
	IsError         string          `json:"isError"` //必须等于0
	ErrCode         string          `json:"errCode"`
}
