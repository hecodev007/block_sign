package bo

type InsertRequest struct {
	UserId  int64  `json:"uid"`
	Address string `json:"address"`
	Url     string `json:"url"`
}

type RemoveRequest struct {
	UserId  int64  `json:"uid"`
	Address string `json:"address"`
}

type RePushRequest struct {
	UserId int64  `json:"uid"`
	Txid   string `json:"txid"`
}

type UpdateRequest struct {
	UserId int64  `json:"uid"`
	Url    string `json:"url"`
}

type InsertContractRequest struct {
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
	Decimal         int    `json:"decimal"`
	CoinType        string `json:"coin_type"`
}

type RemoveContractRequest struct {
	ContractAddresses []string `json:"contract_addresses"`
}
