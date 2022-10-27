package transfer

type AlgoOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Value       string `json:"value"`
	Fee         string `json:"fee"`
	Assert      string `json:"assert"` // 合约地址 主链默认是0
}
