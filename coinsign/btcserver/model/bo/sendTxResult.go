package bo

//远程广播结果返回
type SendTxResult struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    SendTxResultData `json:"data"`
}

type SendTxResultData struct {
	Txid string `json:"txid"`
}
