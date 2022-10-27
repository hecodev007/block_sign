package model

type InsertAddressReq struct {
	Name      string   `json:"name"`
	UserId    int64    `json:"uid"`
	Url       string   `json:"url"`
	Addresses []string `json:"addresses"`
}

type InsertContractReq struct {
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
	Decimal         int    `json:"decimal"`
	CoinType        string `json:"coin_type"`
}

type RemoveRequest struct {
	Name      string   `json:"name"`
	UserId    int64    `json:"uid"`
	Addresses []string `json:"addresses"`
}

type RemoveContractRequest struct {
	Name            string `json:"name"`
	CoinType        string `json:"coin_type"`
	ContractAddress string `json:"contract_address"`
}

type UpdateRequest struct {
	Name   string `json:"name"`
	UserId int64  `json:"uid"`
	Url    string `json:"url"`
}

type RePushRequest struct {
	Name   string `json:"name"`
	UserId int64  `json:"uid"`
	Txid   string `json:"txid"`
	Height int64  `json:"height"`
}

//============================================//
//请求接口参数
type InsertWatchAddressSrvReq struct {
	UserId  int64  `json:"uid"`
	Url     string `json:"url"`
	Address string `json:"address"`
}
