package btcrpc

import "github.com/shopspring/decimal"

//=====================================================chain============================================================
type BtcChaininfo struct {
	Result struct {
		Chain         string `json:"chain"`
		Blocks        int64  `json:"blocks"`
		Headers       int64  `json:"headers"`
		Bestblockhash string `json:"bestblockhash"`
	} `json:"result"`
	Id    int64       `json:"id"`
	Error interface{} `json:"error,omitempty"`
}

//=====================================================chain============================================================

//=====================================================getblockheader============================================================
type BtcBlockHeader struct {
	Result struct {
		Hash              string          `json:"hash"`
		Confirmations     int64           `json:"confirmations"`
		Height            int64           `json:"height"`
		Version           int64           `json:"version"`
		VersionHex        string          `json:"versionHex"`
		Merkleroot        string          `json:"merkleroot"`
		Time              int64           `json:"time"`
		Mediantime        int64           `json:"mediantime"`
		Nonce             int64           `json:"nonce"`
		Bits              string          `json:"bits"`
		Difficulty        decimal.Decimal `json:"difficulty"`
		Chainwork         string          `json:"chainwork"`
		NTx               int             `json:"nTx"` //交易数量
		Previousblockhash string          `json:"previousblockhash"`
		Nextblockhash     string          `json:"nextblockhash"`
	} `json:"result"`
	Id    int64       `json:"id"`
	Error interface{} `json:"error,omitempty"`
}

//=====================================================getblockheader============================================================

//=====================================================BlockCount============================================================
type BtcBlockCountInfo struct {
	Result int64
	Id     int64       `json:"id"`
	Error  interface{} `json:"error,omitempty"`
}

//=====================================================BlockCount============================================================

//===================================================block hash=========================================================

type BtcGetBlockHash struct {
	Result string      `json:"result"`
	Id     int64       `json:"id"`
	Error  interface{} `json:"error,omitempty"`
}

//===================================================block hash=========================================================

//========================================================block=========================================================
type BtcBlock struct {
	Result *BtcBlockInfo `json:"result"`
	Id     int64         `json:"id"`
	Error  interface{}   `json:"error,omitempty"`
}

type BtcBlockInfo struct {
	Hash              string          `json:"hash"`
	Confirmations     int64           `json:"confirmations"`
	Strippedsize      int64           `json:"strippedsize"`
	Size              int64           `json:"size"`
	Weight            int64           `json:"weight"`
	Height            int64           `json:"height"`
	Version           int64           `json:"version"`
	VersionHex        string          `json:"versionHex"`
	Merkleroot        string          `json:"merkleroot"`
	Tx                []*BtcTxInfo    `json:"tx"`
	Time              int64           `json:"time"`
	Mediantime        int64           `json:"mediantime"`
	Nonce             int64           `json:"nonce"`
	Bits              string          `json:"bits"`
	Difficulty        decimal.Decimal `json:"difficulty"`
	Chainwork         string          `json:"chainwork"`
	NTx               int             `json:"nTx"` //交易数量
	Previousblockhash string          `json:"previousblockhash"`
	Nextblockhash     string          `json:"nextblockhash"`
}

type BtcBlock1 struct {
	Result *BtcBlockInfo1 `json:"result"`
	Id     int64          `json:"id"`
	Error  interface{}    `json:"error,omitempty"`
}

type BtcBlockInfo1 struct {
	Hash              string          `json:"hash"`
	Confirmations     int64           `json:"confirmations"`
	Strippedsize      int64           `json:"strippedsize"`
	Size              int64           `json:"size"`
	Weight            int64           `json:"weight"`
	Height            int64           `json:"height"`
	Version           int64           `json:"version"`
	VersionHex        string          `json:"versionHex"`
	Merkleroot        string          `json:"merkleroot"`
	Tx                []string        `json:"tx"`
	Time              int64           `json:"time"`
	Mediantime        int64           `json:"mediantime"`
	Nonce             int64           `json:"nonce"`
	Bits              string          `json:"bits"`
	Difficulty        decimal.Decimal `json:"difficulty"`
	Chainwork         string          `json:"chainwork"`
	NTx               int             `json:"nTx"` //交易数量
	Previousblockhash string          `json:"previousblockhash"`
	Nextblockhash     string          `json:"nextblockhash"`
}

//========================================================block=========================================================

//================================================= tx==================================================================
type BtcTx struct {
	Result *BtcTxInfo  `json:"result"`
	Id     int64       `json:"id"`
	Error  interface{} `json:"error,omitempty"`
}

type BtcTxInfo struct {
	Txid          string             `json:"txid"`
	Hash          string             `json:"hash"`
	Size          int64              `json:"size"`
	Vsize         int64              `json:"vsize"`
	Version       int                `json:"version"`
	Locktime      int64              `json:"locktime"`
	Vin           []*BtcVin          `json:"vin"`
	Vout          []*BtcVout         `json:"vout"`
	Blockhash     string             `json:"blockhash"`
	Confirmations int64              `json:"confirmations"`
	Time          int64              `json:"time"`
	Blocktime     int64              `json:"blocktime"`
	Hex           string             `json:"hex"`
	Contract      []*OmniTransaction `json:"contract,omitempty"`
	Fee           decimal.Decimal    //额外补充业务逻辑需要的手续费
	VinAmount     decimal.Decimal    //额外补充业务逻辑需要的vin总输入
	VoutAmount    decimal.Decimal    //额外补充业务逻辑需要的vout总输入
}

type BtcVin struct {
	Sequence  int64           `json:"sequence"`
	Coinbase  string          `json:"coinbase,omitempty"`
	Txid      string          `json:"txid,omitempty"`
	Vout      int             `json:"vout,omitempty"`
	ScriptSig BtcScriptSig    `json:"scriptSig,omitempty"`
	Address   string          `json:"address"` //额外在业务补充
	Amount    decimal.Decimal `json:"amount"`  //额外在业务补充
}

type BtcScriptSig struct {
	Asm string `json:"asm,omitempty"`
	Hex string `json:"hex,omitempty"`
}

type BtcVout struct {
	Value        decimal.Decimal `json:"value"`
	N            int             `json:"n"`
	ScriptPubKey BtcScriptPubKey `json:"scriptPubKey"`
}

type BtcScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

//================================================= tx==================================================================

//================================================= usdt==================================================================

type OmniGettransaction struct {
	Result *OmniTransaction `json:"result"`
	Id     int64            `json:"id"`
	Error  interface{}      `json:"error,omitempty"`
}

type OmniTransaction struct {
	Txid             string          `json:"txid"`
	Fee              decimal.Decimal `json:"fee"`
	Sendingaddress   string          `json:"sendingaddress"`
	Referenceaddress string          `json:"referenceaddress"`
	Ismine           bool            `json:"ismine"`
	Version          int             `json:"version"`
	TypeInt          int             `json:"type_int"`
	Type             string          `json:"type"`
	Propertyid       int             `json:"propertyid"`
	Divisible        bool            `json:"divisible"`
	Amount           decimal.Decimal `json:"amount"`
	Valid            bool            `json:"valid"`
	Blockhash        string          `json:"blockhash"`
	Blocktime        int64           `json:"blocktime"`
	Positioninblock  int             `json:"positioninblock"`
	Block            int64           `json:"block"`
	Confirmations    int64           `json:"confirmations"`
}

//================================================= usdt==================================================================
