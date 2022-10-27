// getblockhash resp
package bo

// push type
type PushType int32

const (
	// utxo
	PushTypeTX     PushType = 0 // 交易数据
	PushTypeConfir PushType = 1 // 确认数更新

	// account
	PushTypeAccountTX     PushType = 10 // 交易数据
	PushTypeAccountConfir PushType = 11 // 确认数更新

	PushTypeAtomTX     PushType = 40 // 交易数据
	PushTypeAtomConfir PushType = 41 // 确认数更新
)

// 获取指定高度的hash
type GetBlockHashResult struct {
	Id     string `json:"id"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// 获取块高度
type GetBlockCountResult struct {
	Id     string `json:"id"`
	Result int64  `json:"result"`
	Error  string `json:"error"`
}

// 块数据
type GetBlockResult struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	Size              int64    `json:"size"`
	Height            int64    `json:"height"`
	Version           int64    `json:"version"`
	Time              int64    `json:"time"`
	Previousblockhash string   `json:"previousblockhash"`
	Nextblockhash     string   `json:"nextblockhash"`
	Tx                []string `json:"tx"`
}

// input签名
type TxInputScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// output 公钥
type TxOutputScriptPutKey struct {
	Addresse string `json:"address"`
}

// Input
type TxInput struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

// Output
type TxOutput struct {
	Addresse string `json:"address"`
	Value    int64  `json:"value"`
	N        int    `json:"n"`
}

// 交易数据
type TxData struct {
	Txid     string   `json:"txid"`
	Hash     string   `json:"hash"`
	Version  int64    `json:"version"`
	Size     int64    `json:"size"`
	Vsize    int64    `json:"vsize"`
	Locktime int64    `json:"locktime"`
	Vin      TxInput  `json:"vin"`
	Vout     TxOutput `json:"vout"`
}

// 地址信息
type UserAddressInfo struct {
	UserID    int64
	Address   string
	NotifyUrl string
	AccountID string
}

// Input
type PushTxInput struct {
	Txid     string `json:"txid"`
	Vout     int    `json:"vout"`
	Addresse string `json:"address"`
	Value    string `json:"value"`
}

// Output
type PushTxOutput struct {
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
	Txid     string           `json:"txid"`
	Fee      string           `json:"fee"`
	Coinbase bool             `json:"iscoinbase"`
	Vin      []PushTxInput    `json:"vin"`
	Vout     []PushTxOutput   `json:"vout"`
	Contract []PushContractTx `json:"contract"`
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
	Type          PushType        `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string          `json:"coin"`
	Token         string          `json:"token"`
	Height        int64           `json:"height"`
	Hash          string          `json:"hash"`
	Confirmations int64           `json:"confirmations"`
	Time          int64           `json:"time"`
	Txs           []PushAccountTx `json:"txs"`
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

// push block
type PushAtomBlockInfo struct {
	Type          PushType     `json:"type"` // 0:推送交易数据，1：推送块确认数更新
	CoinName      string       `json:"coin"`
	Height        int64        `json:"height"`
	Hash          string       `json:"hash"`
	Confirmations int64        `json:"confirmations"`
	Time          int64        `json:"time"`
	Txs           []PushAtomTx `json:"txs"`
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
