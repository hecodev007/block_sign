package domain

type SearchChains struct {
	Id         int64  `json:"id"`
	Account    string `json:"account"`
	MerchantId int64  `json:"merchant_id"`
	ServiceId  int    `json:"service_id"`
	Limit      int    `json:"limit"  description:"查询条数" example:"10"`
	Offset     int    `json:"offset" description:"查询起始位置" example:"0"`
}

type UpdateChains struct {
	Id           int64  `json:"id"`
	MerchantId   int64  `json:"merchant_id"`
	Account      string `json:"account"`
	ServiceId    int    `json:"service_id,omitempty"`
	CoinId       int    `json:"coin_id,omitempty"`
	ChainAddr    string `json:"chain_addr,omitempty"`
	IsGetAddr    int    `json:"is_get_addr,omitempty"`
	IsWithdrawal int    `json:"is_withdrawal,omitempty"`
	Reason       string `json:"reason"`
	IsList       []int  `json:"is_list"`
}

type MerchantServiceChains struct {
	Serial      int    `json:"serial"`
	ServiceId   int    `json:"service_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ChainName   string `json:"chain_name"`
	CoinName    string `json:"coin_name"`
	AdminNums   int64  `json:"admin_nums"`
	AuditNums   int64  `json:"audit_nums"`
	FinanceNums int64  `json:"finance_nums"`
	VisitorNums int64  `json:"visitor_nums"`
}

type ServiceAndCoin struct {
	Id             int64         `json:"id,omitempty"`
	ServiceId      int           `json:"service_id,omitempty"`
	ServiceName    string        `json:"service_name,omitempty"`
	ChainsCoinList []ChainsCoins `json:"chains_coin_list"`
}

type ChainsCoins struct {
	ChainName string `json:"chain_name"`
	CoinName  string `json:"coin_name"`
}

type ServiceRolesInfo struct {
	ServiceId   int          `json:"service_id,omitempty"`
	ServiceName string       `json:"service_name,omitempty"`
	AdminInfo   ServiceRoles `json:"admin_info"`
	AuditInfo   ServiceRoles `json:"audit_info"`
	FinanceInfo ServiceRoles `json:"finance_info"`
	VisitorInfo ServiceRoles `json:"visitor_info"`
}

type ServiceRoles struct {
	Name      string   `json:"name"`
	Nums      int      `json:"nums"`
	UserAndId []string `json:"user_and_id"`
}
