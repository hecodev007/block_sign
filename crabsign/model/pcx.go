package model

type PcxTransferParams struct {
	PcxReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Memo        string `json:"memo"`
	Token       string `json:"token"`
	//cold sign need
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
}
type PcxReqBaseParams struct {
	OrderId    string `json:"order_no,omitempty"`
	MchId      int    `json:"mch_id,omitempty"`
	CoinName   string `json:"coin_name,omitempty"`
	Sign       string `json:"sign"`
	CreateTime string `json:"createTime"` // time.Now().Unix()  长度为10位的时间戳
}
type PcxColdParams struct {
	ReqBaseParams
	FromAddress        string `json:"from_address"`
	ToAddress          string `json:"to_address"`
	Amount             string `json:"amount"`
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
	GenesisHash        string `json:"genesis_hash"`
	BlockHash          string `json:"block_hash"`
	BlockNumber        uint64 `json:"block_number"`
}
