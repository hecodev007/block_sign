package bo

type TransferRequest struct {
	ApplyId      int64  `json:"applyid"`      //商户ID
	OuterOrderNo string `json:"outerorderno"` //外部订单号
	OrderNo      string `json:"orderno"`      //内部订单号
	MchName      string `json:"mchname"`      //商户名称
	CoinName     string `json:"coinname"`     //币种名称
	FromAddress    string `json:"fromaddress"`    //发送地址
	ToAddress    string `json:"toaddress"`    //接收地址
	ToAmount     string `json:"toamount"`     //接收金额
	Memo         string `json:"memo"`         //memo
}

type TransferRpcResponse struct {
	JsonRpc string            `json:"jsonrpc"`
	Id      int               `json:"id"`
	Result  []interface{}     `json:"result"`
	Error   *TransferRpcError `json:"error"`
}

type TransferRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TransferReturn struct {
	Code int
	TRR  TransferRpcResponse
}

type CreateAccountReq struct {
	Account    string    `json:"account"`
}
