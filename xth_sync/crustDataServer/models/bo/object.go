// getblockhash resp
package bo

// push type
type PushType int32

const (
	// utxo
	PushTypeTX     PushType = 0 // 交易数据
	PushTypeConfir PushType = 1 // 确认数更新

)

// 地址信息
type UserAddressInfo struct {
	UserID    int64
	Address   string
	NotifyUrl string
	AccountID string
}

// Input
type PushTxInput struct {
	AssetID  string `json:"assetID"`
	Txid     string `json:"txid"`
	Vout     int    `json:"vout"`
	Addresse string `json:"address"`
	Value    string `json:"value"`
}

// Output
type PushTxOutput struct {
	AssetID  string `json:"assetID"`
	Addresse string `json:"address"`
	Value    string `json:"value"`
	N        int    `json:"n"`
}

// Output
type PushContractTx struct {
	Contract string `json:"contract"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Fee      string `json:"fee"`
	MaxFee   string `json:"maxfee"`
}

// 交易信息
type PushUtxoTx struct {
	Txid     string            `json:"txid"`
	Fee      string            `json:"fee"`
	Coinbase bool              `json:"iscoinbase"`
	Vin      []*PushTxInput    `json:"vin,omitempty"`
	Vout     []*PushTxOutput   `json:"vout,omitempty"`
	Contract []*PushContractTx `json:"contract,omitempty"`
}

// push block
type PushUtxoBlockInfo struct {
	Type          PushType      `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string        `json:"coin"`
	Height        int64         `json:"height"`
	Hash          string        `json:"hash"`
	Confirmations int64         `json:"confirmations"`
	Time          int64         `json:"time"`
	Txs           []*PushUtxoTx `json:"txs"`
}

// 交易信息
type PushAccountTx struct {
	Name     string `json:"name"`
	Txid     string `json:"txid"`
	Fee      string `json:"fee"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Memo     string `json:"memo"`
	Contract string `json:"contract"`
}

// push block
type PushAccountBlockInfo struct {
	Type          PushType         `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string           `json:"coin"`
	Token         string           `json:"token"`
	Height        int64            `json:"height"`
	Hash          string           `json:"hash"`
	Confirmations int64            `json:"confirmations"`
	Time          int64            `json:"time"`
	Txs           []*PushAccountTx `json:"txs"`
}

type PushAtomTxMsg struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
}

type PushAtomAccountTxMsg struct {
	PushAtomTxMsg
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type AtomInput struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type AtomOutput AtomInput

type PushAtomUtxoTxMsg struct {
	PushAtomTxMsg
	Inputs  []AtomInput  `json:"inputs"`
	Outputs []AtomOutput `json:"outputs"`
}

// 交易信息
type PushAtomTx struct {
	Txid   string        `json:"txid"`
	Fee    string        `json:"fee"`
	Memo   string        `json:"memo"`
	TxMsgs []interface{} `json:"txmsgs"`
}

// 推送结果
type PushResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// cocos head block
type CocosHeadBlock struct {
	Height int64  `json:"head_block_number"`
	Hash   string `json:"head_block_id"`
}

// 合约信息
type ContractInfo struct {
	Name            string `json:"name"`            // 合约名称
	ContractAddress string `json:"contractaddress"` // 合约地址
	Decimal         int    `json:"decimal"`         // 精度
}
