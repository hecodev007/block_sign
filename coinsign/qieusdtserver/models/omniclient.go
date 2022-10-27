package models

//================params================

type OmniCreateTx struct {
	TxIns []OmniSimpleTxin
	Out   map[string]interface{}
}

type OmniSimpleTxin struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

type OmniChangeTx struct {
	Rawtx       string                `json:"rawtx"`       //required	the raw transaction to extend
	Prevtxs     []OmniChangeTxPrevtxs `json:"prevtxs"`     //required	a JSON array of transaction inputs ---->[]RawTxPrevtxs 数组json
	Destination string                `json:"destination"` //required	the destination for the change
	Fee         float64               `json:"fee"`         //number	required	the desired transaction fees
	//Position    int                   `json:"position,omitempty"` //number	optional	the position of the change output (default: first position)
}

type OmniChangeTxPrevtxs struct {
	Txid         string  `json:"txid"`         // (string, required) the transaction hash
	Vout         int     `json:"vout"`         // (number, required) the output number
	ScriptPubKey string  `json:"scriptPubKey"` // (string, required) the output script
	Value        float64 `json:"value"`        // (number, required) the output value
	RedeemScript string  `json:"redeemScript"` //(string, required for P2SH or P2WSH) redeem script
}

type OmniSigntx struct {
	Rawtx   string              `json:"rawtx"`   //required	the raw transaction to extend
	Prevtxs []OmniSigntxPrevtxs `json:"prevtxs"` //required	a JSON array of transaction inputs ---->[]RawTxPrevtxs 数组json
	Prvkey  []string            `json:"prvkey"`  //冗余字段测试
}

type OmniSigntxPrevtxs struct {
	Txid         string  `json:"txid"`         // (string, required) the transaction hash
	Vout         int     `json:"vout"`         // (number, required) the output number
	ScriptPubKey string  `json:"scriptPubKey"` // (string, required) the output script
	Value        float64 `json:"value"`        // (number, required) the output value
	RedeemScript string  `json:"redeemScript"` //(string, required for P2SH or P2WSH) redeem script
}

//================result================
//构造简单发送金额返回结果
type OmniSimpleSendResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

//创建交易放回结果
type OmniCreaterawtransactionResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

type OmniOpreturnResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

type OmniReferenceResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

type OmniChangeResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

type OmniSignResult struct {
	Result struct {
		Hex      string      `json:"hex"`
		Complete bool        `json:"complete"`
		Errors   interface{} `json:"errors,omitempty"`
	} `json:"result"` //返回结果
	Error interface{} `json:"error"` //错误结果
	Id    uint64      `json:"id"`
}

//获取新地址
type OmniGetNewAddressResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

//获取地址私钥
type OmniDumpprivkeyResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

//获取地址私钥
type OmniImportprivkeyResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

//获取地址公钥
type OmniValidateaddressResult struct {
	Result OmniValidateResultData `json:"result"` //返回结果
	Error  *ErrContext            `json:"error"`  //错误结果
	Id     uint64                 `json:"id"`
}

type OmniValidateResultData struct {
	Address      string `json:"address"`      //地址
	ScriptPubKey string `json:"scriptPubKey"` //脚本公钥
	Pubkey       string `json:"pubkey"`       //公钥
	Account      string `json:"account"`      //账号
}

type OmniSenndTxResult struct {
	Result string      `json:"result"` //返回结果
	Error  *ErrContext `json:"error"`  //错误结果
	Id     uint64      `json:"id"`
}

//==============rpc导入address==================

type ImportaddrResult struct {
	Result string `json:"result"` //返回结果
	Error  struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"` //错误结果
	Id uint64 `json:"id"`
}

//==============rpc导入address==================
//==============rpc获取unspent==================
type ListunspentResult struct {
	Result []RpcUnspent `json:"result"` //返回结果
	Error  *ErrContext  `json:"error"`  //错误结果
	Id     uint64       `json:"id"`
}

type RpcUnspent struct {
	Txid          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Confirmations int64   `json:"confirmations"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
}

//==============rpc获取unspent==================

type ErrContext struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
