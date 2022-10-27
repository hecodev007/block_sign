package model

// ========================create address==================//
type ReqBaseParams struct {
	OrderId  string `json:"orderId,omitempty"`
	MchId    string `json:"mchId,omitempty"`
	CoinName string `json:"coinName,omitempty"`
}
type ReqCreateAddressParamsV2 struct {
	// 商户名称,如：hoo
	MchId string `json:"mch,omitempty"`
	// 币种，如：trx
	CoinName string `json:"coinCode,omitempty"`
	// 生成地址数量
	Num int `json:"count,omitempty"`
	// 本次生成地址对应的编号，如：trx_usb_20220601001
	OrderId string `json:"batchNo,omitempty"`
}

//type ReqCreateAddressParamsV2 struct {
//	MchId    string `json:"mchId"`
//	CoinName string `json:"coinName"`
//	Num      int    `json:"num"`
//	OrderId  string `json:"orderId"`
//}
type ReqCreateAddressParams struct {
	Num int `json:"num"`
	ReqBaseParams
}

type RespCreateAddressParams struct {
	// 商户名称,如：hoo（直接使用入参的值）
	Mch string `json:"mch,omitempty"`
	// 币种，如：trx（直接使用入参的值）
	CoinCode string `json:"coinCode,omitempty"`
	// 本次生成地址对应的编号，如：trx_usb_20220601001（直接使用入参的值）
	BatchNo string `json:"batchNo,omitempty"`
	// 本次生成的地址列表
	Address []string `json:"address"`
}

//type RespCreateAddressParams struct {
//	ReqCreateAddressParams
//	Address []string `json:"address"`
//}

// ==========================================================//

type ReqSignParams struct {
	ReqBaseParams
	Data interface{} `json:"data"`
}

type RespSignParams struct {
	ReqBaseParams
	Result string `json:"result"`
}

// =============================================================//

type ReqGetBalanceParams struct {
	CoinName        string      `json:"coin_name"`        // 币种主链的名字
	Address         string      `json:"address"`          //	需要获取余额的地址
	Token           string      `json:"token"`            // 	token的名字
	ContractAddress string      `json:"contract_address"` // 合约地址
	Params          interface{} `json:"params"`           // 特殊参数（如果有特殊参数，传入到这里面）
}

// ------------------valid address-------------------
type ReqValidAddressParams struct {
	Address string `json:"address"`
}
