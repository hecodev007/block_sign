package model

import "github.com/shopspring/decimal"

// push type
type PushType int32

const (
	// utxo
	PushTypeTX     PushType = 0 // 交易数据
	PushTypeConfir PushType = 1 // 确认数更新

	// account
	PushTypeAccountTX     PushType = 10 // 交易数据
	PushTypeAccountConfir PushType = 11 // 确认数更新

	// 分叉
	PushTypeChainFork PushType = 20 // 区块分叉

	// btm 特殊
	BtmPushTypeTX     PushType = 30 // BTM交易数据
	BtmPushTypeConfir PushType = 31 // BTM确认数更新

	// cosmos
	PushTypeAtomTX     PushType = 40 // 交易数据
	PushTypeAtomConfir PushType = 41 // 确认数更新

	// eos
	PushTypeEosTX PushType = 50 // eos交易数据
)

// push block, 不含交易基础结构
type PushBaseBlockInfo struct {
	Type          PushType `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string   `json:"coin"`
	Height        int64    `json:"height"`
	Hash          string   `json:"hash"`
	Confirmations int64    `json:"confirmations"`
	Time          int64    `json:"time"`
}

// ///////////////////////////////////////////////////////////////////////////
// utxo模型

type AssetsInfo struct {
	Name       string `json:"name"`
	AssetId    string `json:"id"`
	AssetValue string `json:"value"` // token 金额
}

// Input
type PushTxInput struct {
	Txid     string      `json:"txid"`
	Vout     int         `json:"vout"`
	Addresse string      `json:"address"`
	Value    string      `json:"value"`
	Assets   *AssetsInfo `json:"assets,omitempty"`

	// neo 使用
	AssetName string `json:"assetname,omitempty"`
	AssetId   string `json:"assetid,omitempty"`
}

// Output
type PushTxOutput struct {
	Addresse    string      `json:"address"`
	RawAddresse string      `json:"rawAddress,omitempty"`
	Value       string      `json:"value"`
	N           int         `json:"n"`
	Assets      *AssetsInfo `json:"assets,omitempty"`

	// neo 使用
	AssetName string `json:"assetname,omitempty"`
	AssetId   string `json:"assetid,omitempty"`

	// ckb使用
	CodeHash string `json:"codehash,omitempty"`
}

// Output
type PushContractTx struct {
	Contract string      `json:"contract"`
	From     string      `json:"from"`
	To       string      `json:"to"`
	Amount   string      `json:"amount"`
	Fee      interface{} `json:"fee"`
	MaxFee   interface{} `json:"maxfee"`
	Memo     string      `json:"memo"`
	Coin     string      `json:"coin"`
	FeeCoin  string      `json:"feeCoin"`
	Valid    bool        `json:"valid"`
}

// 交易信息
type PushUtxoTx struct {
	Txid      string           `json:"txid"`
	Fee       interface{}      `json:"fee"`
	Coinbase  bool             `json:"iscoinbase"`  // 挖矿
	CoinStake bool             `json:"iscoinstake"` // qtum gas
	Vin       []PushTxInput    `json:"vin"`
	Vout      []PushTxOutput   `json:"vout"`
	Contract  []PushContractTx `json:"contract,omitempty"`
}

// push block
type PushUtxoBlockInfo struct {
	Type          PushType     `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string       `json:"coin"`
	Height        int64        `json:"height"`
	Hash          string       `json:"hash"`
	Confirmations int64        `json:"confirmations"`
	Time          int64        `json:"time"`
	Txs           []PushUtxoTx `json:"txs"`
}

// ///////////////////////////////////////////////////////////////////////////
// 账号模型

// 交易信息
type PushAccountTx struct {
	Txid        string      `json:"txid"`
	Fee         interface{} `json:"fee"`
	From        string      `json:"from"`
	To          string      `json:"to"`
	Amount      string      `json:"amount"`
	Memo        string      `json:"memo"`
	MemoEncrypt string      `json:"memo_encrypt"`
	Contract    string      `json:"contract"`
	FeePayer    string      `json:"feepayer"`

	// 以下三个为 入金 风控新增的字段 2021.06.24
	IsRisk    bool   `json:"isrisk"`
	RiskLevel int    `json:"risklevel"`
	RiskMsg   string `json:"riskmsg"`

	// 目前hx使用  固定为1.3.0
	Assetid string `json:"assetid"`
}

// push block
type PushAccountBlockInfo struct {
	Type          PushType        `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string          `json:"coin"`
	Height        int64           `json:"height"`
	Hash          string          `json:"hash"`
	Confirmations int64           `json:"confirmations"`
	Time          int64           `json:"time"`
	Txs           []PushAccountTx `json:"txs"`
}

// ///////////////////////////////////////////////////////////////////////////
// eos,fibos模型

// 交易信息
type PushEosAction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Memo     string `json:"memo"`
	Contract string `json:"contract"`
	Token    string `json:"token"`
}

// 交易信息
type PushEosTx struct {
	Txid    string          `json:"txid"`
	Fee     interface{}     `json:"fee"`
	Status  string          `json:"status"`
	Actions []PushEosAction `json:"actions"`
}

// push block
type PushEosBlockInfo struct {
	Type          PushType    `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string      `json:"coin"`
	Height        int64       `json:"height"`
	Hash          string      `json:"hash"`
	Confirmations int64       `json:"confirmations"`
	Time          int64       `json:"time"`
	Txs           []PushEosTx `json:"txs"`
}

// btm

type PushBtmBlockInfo struct {
	Type          PushType    `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string      `json:"coin"`
	Height        int64       `json:"height"`
	Hash          string      `json:"hash"`
	Confirmations int64       `json:"confirmations"`
	Time          int64       `json:"time"`
	Txs           []PushBtmTx `json:"txs"`
}

// 交易信息
type PushBtmTx struct {
	Txid       string          `json:"txid"`
	FeeFloat   decimal.Decimal `json:"fee"`
	Coinbase   bool            `json:"iscoinbase"` // 挖矿
	StatusFail bool            `json:"statusFail"` // 是否是失败交易
	Vin        []PushBtmInput  `json:"vin"`
	Vout       []PushBtmOutput `json:"vout"`
	MuxId      string          `json:"muxId"`
}

type BtmTxType string

var (
	BtmSpend   BtmTxType = "spend"
	BtmControl BtmTxType = "control"
)

// Input
type PushBtmInput struct {
	Type           BtmTxType       `json:"type"` // 固定spend
	SpentOutputId  string          `json:"spentOutputId"`
	AssetId        string          `json:"assetId"`
	AssetName      string          `json:"assetName"`
	AmountFloat    decimal.Decimal `json:"amount"`
	Address        string          `json:"address"`
	InputId        string          `json:"inputId"`
	ControlProgram string          `json:"controlProgram"`
}

// Output
type PushBtmOutput struct {
	Type           BtmTxType       `json:"type"` // 固定control
	VoutId         string          `json:"voutId"`
	Position       int             `json:"position"`
	AssetId        string          `json:"assetId"`
	AssetName      string          `json:"assetName"`
	AmountFloat    decimal.Decimal `json:"amount"`
	Address        string          `json:"address"`
	InputId        string          `json:"inputId"`
	ControlProgram string          `json:"controlProgram"`
}
