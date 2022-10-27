package bo

type BrainKey struct {
	BrainPrivKey string `json:"brain_priv_key"`
	WifPrivKey   string `json:"wif_priv_key"`
	PubKey       string `json:"pub_key"`
}

type RegisterResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}
