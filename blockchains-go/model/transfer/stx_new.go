package transfer

type StxNewOrderRequest struct {
	OrderRequestHead

	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Value       string `json:"value"`
	Fee         string `json:"fee"`
	Memo        string `json:"memo"`
	Nonce       int64  `json:"nonce"`
	//  token 转账 需要参数
	ContractAddress string `json:"contract_address"`
	Token           string `json:"token"` // 代币的名字，主链转账不传这个值

}
