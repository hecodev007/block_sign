package avax

func (rpc *RpcClient) SendRawTransaction(rawTx string) (txid string, err error) {

	param := struct {
		Tx string `json:"tx"`
	}{Tx: rawTx}
	response := struct {
		TxID string `json:"txId"`
	}{}
	err = rpc.CallWithAuth("/ext/bc/X", "avm.issueTx", rpc.Credentials, &response, param)
	return response.TxID, err
}
