package bo

type BalanceRpcResponse struct {
	JsonRpc string              `json:"jsonrpc"`
	Id      int                 `json:"id"`
	Result  []*BalanceRpcResult `json:"result"`
	Error   *BalanceRpcError    `json:"error"`
}

type BalanceRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BalanceRpcResult struct {
	Amount  interface{} `json:"amount"`
	AssetId string `json:"asset_id"`
}

type BalanceReturn struct {
	Code int
	BRR  BalanceRpcResponse
}
