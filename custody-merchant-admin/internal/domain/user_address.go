package domain

type InsertAddrInfo struct {
	Id           int64  `json:"id"`
	MerchantId   int64  `json:"merchant_id"`
	MerchantUser string `json:"merchant_user"`
	ServiceId    int64  `json:"service_id"`
	CoinId       int    `json:"coin_id"`
	ChainId      int    `json:"chain_id"`
	Address      string `json:"address"`
	State        int    `json:"state"`
	ClientId     string `json:"client_id"`
	SecureKey    string `json:"secure_key"`
}

type GetAddrInfo struct {
	UserId   string `json:"user_id"`
	Coin     string `json:"coin"`
	Chain    string `json:"chain"`
	ClientId string `json:"clint_id"`
	Sign     string `json:"sign"`
	Nonce    string `json:"nonce"`
	Ts       int64  `json:"ts"`
}

type BatchAddrInfo struct {
	UserId   []string `json:"user_id"`
	Nums     int      `json:"nums"`
	Coin     string   `json:"coin"`
	Chain    string   `json:"chain"`
	ClientId string   `json:"client_id"`
	Sign     string   `json:"sign"`
	Nonce    string   `json:"nonce"`
	Ts       int64    `json:"ts"`
}

type GetBillInfo struct {
	SerialNo  string `json:"serial_no"`
	ClientId  string `json:"client_id"`
	SecureKey string `json:"secureKey"`
	Sign      string `json:"sign"`
}

type BatchAddrStruct struct {
	UserId   []string `json:"user_id" form:"user_id"`
	Coin     string   `json:"coin" form:"coin"`
	Chain    string   `json:"chain" form:"chain"`
	ClientId string   `json:"client_id" form:"client_id"`
	Sign     string   `json:"sign" form:"sign"`
	Nonce    string   `json:"nonce" form:"nonce"`
	Ts       float64  `json:"ts" form:"ts"`
}

type BatchAddrParam struct {
	UserId   []string `json:"user_id" form:"user_id"`
	Nums     int      `json:"nums" form:"nums"`
	Coin     string   `json:"coin" form:"coin"`
	Chain    string   `json:"chain" form:"chain"`
	ClientId string   `json:"client_id" form:"client_id"`
	Sign     string   `json:"sign" form:"sign"`
	Nonce    string   `json:"nonce" form:"nonce"`
	Ts       int64    `json:"ts" form:"ts"`
}

func BatchAddrTo(a *BatchAddrStruct) *BatchAddrInfo {
	return &BatchAddrInfo{
		UserId:   a.UserId,
		Coin:     a.Coin,
		Chain:    a.Chain,
		ClientId: a.ClientId,
		Sign:     a.Sign,
		Nonce:    a.Nonce,
		Ts:       int64(a.Ts),
	}
}
