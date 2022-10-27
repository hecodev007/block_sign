package bo

type RpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RpcError   `json:"error"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
