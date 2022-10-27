package models

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

// 推送结果
type PushResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
