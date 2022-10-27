// getblockhash resp
package models

// 链同步服务动作
type ChainCoreAction interface {
	InitWatchAddress() // 初始化关注地址
	InitPush()         // 初始化推送
	InitContract()     // 初始化合约
	InitSync()         // 初始化同步服务
	RouterInit()       // 初始化路由

	RunPush()   // 运行推送服务
	StartSync() // 运行同步服务
}

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
	AssetName *string `json:"assetname,omitempty"`
	AssetId   *string `json:"assetid,omitempty"`
}

// Output
type PushTxOutput struct {
	Addresse string      `json:"address"`
	Value    string      `json:"value"`
	N        int         `json:"n"`
	Assets   *AssetsInfo `json:"assets,omitempty"`

	// neo 使用
	AssetName *string `json:"assetname,omitempty"`
	AssetId   *string `json:"assetid,omitempty"`

	// ckb使用
	CodeHash *string `json:"codehash,omitempty"`
}

// Output
type PushContractTx struct {
	Contract string  `json:"contract"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   string  `json:"amount"`
	Fee      float64 `json:"fee"`
	MaxFee   float64 `json:"maxfee"`
}

// 交易信息
type PushUtxoTx struct {
	Txid      string           `json:"txid"`
	Fee       float64          `json:"fee"`
	Coinbase  bool             `json:"iscoinbase"`
	Coinstake bool             `json:"iscoinstake"`
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

// 交易信息
type PushAccountTx struct {
	Txid     string  `json:"txid"`
	Fee      float64 `json:"fee"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   string  `json:"amount"`
	Memo     string  `json:"memo"`
	Contract string  `json:"contract"`
	FeePayer string  `json:"feepayer"`
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
	Fee     float64         `json:"fee"`
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
	ActTransferName string `json:"transfer"`        // 合约交易事件名称
}

// token信息,mtr目前使用
type TokenInfo struct {
	Name    string `json:"name"`    // token名称
	TokenId int64  `json:"tokenid"` // token
	Decimal int    `json:"decimal"` // 精度
}

// 心跳信息
type HeartInfo struct {
	Coin   string `json:"coin"`
	Height int64  `json:"height"`
}
