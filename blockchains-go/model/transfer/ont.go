package transfer

type OntOrderHeadReq struct {
	ApplyId      int64  `json:"applyId,omitempty"`
	ApplyCoinId  int64  `json:"applyCoinId,omitempty"`
	OuterOrderNo string `json:"outerOrderNo,omitempty"`
	OrderNo      string `json:"orderNo,omitempty"`
	MchId        int64  `json:"mchId,omitempty"`
	MchName      string `json:"mchName,omitempty"`
	CoinName     string `json:"coinName,omitempty"`
}
type OntOrderRequest struct {
	OntOrderHeadReq
	//CoinType	string	`json:"coinName"`
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"` // '接收者地址'
	Amount      int64  `json:"amount"`
	GasPrice    int64  `json:"gasPrice"`
	GasLimit    int64  `json:"gasLimit"`
}
