package model

const (
	ResponseCodeFail    = 1
	ResponseCodeSuccess = 0
)

// ========================create address==================//
type ReqBaseParams struct {
	OrderId  string `json:"orderId,omitempty"`
	MchId    string `json:"mchId,omitempty"`
	CoinName string `json:"coinName,omitempty"`
}

type ReqCreateAddressParams struct {
	Num int `json:"num"`
	ReqBaseParams
}

type ReqCreateAddressParamsV2 struct {
	Mch      string `json:"mch,omitempty"`
	CoinCode string `json:"coinCode,omitempty"`
	Count    int    `json:"count,omitempty"`
	BatchNo  string `json:"batchNo,omitempty"`
}

type RespCreateAddressParams struct {
	Mch      string   `json:"mch,omitempty"`
	CoinCode string   `json:"coinCode,omitempty"`
	BatchNo  string   `json:"batchNo,omitempty"`
	Address  []string `json:"address"`
}

type ReqSignParams struct {
	ReqBaseParams
	Data interface{} `json:"data"`
}

type RespSignParams struct {
	ReqBaseParams
	Result string `json:"result"`
}

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

type ReqDelKeyParams struct {
	OrderId string `json:"order_id"`
}

type CallbackReqParams struct {
	Data    CallbackReqParamsData `json:"data"`
	Code    int                   `json:"code"`
	Message string                `json:"message"`
}

type CallbackReqParamsData struct {
	TxId         string `json:"tx_id"`
	OuterOrderNo string `json:"outer_order_no"`
	OrderHotId   int    `json:"order_hot_id"`
	Success      bool   `json:"success"`
}

type ReCallback struct {
	RetryCount int
	SendTime   int64
	ErrMsg     string
	CallbackReqParams
}


type BscTransferParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	IsCollect       int    `json:"is_collect"`
	Token           string `json:"token"`
	ContractAddress string `json:"contract_address"`
	Fee             int64  `json:"fee"`
	GasPrice        int64  `json:"gasPrice"`
	GasLimit        int64  `json:"gasLimit"`
}