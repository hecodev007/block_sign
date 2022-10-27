package rylink

type ErrRPC struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

//获取新地址
type GetNewAddressResult struct {
	Result string `json:"result"` //返回结果
	Error  string `json:"error"`  //错误结果
	Id     uint64 `json:"id"`
}

//获取地址私钥
type DumpprivkeyResult struct {
	Result string `json:"result"` //返回结果
	Error  string `json:"error"`  //错误结果
	Id     uint64 `json:"id"`
}

//创建交易放回结果
type CreaterawtransactionResult struct {
	Result string  `json:"result"` //返回结果
	Error  *ErrRPC `json:"error"`  //错误结果
	Id     uint64  `json:"id"`
}

type SignResult struct {
	Result struct {
		Hex      string      `json:"hex"`
		Complete bool        `json:"complete"`
		Errors   interface{} `json:"errors,omitempty"`
	} `json:"result"` //返回结果
	Error *ErrRPC `json:"error"` //错误结果
	Id    uint64  `json:"id"`
}

type SenndTxResult struct {
	Result string `json:"result"` //返回结果
	Error  string `json:"error"`  //错误结果
	Id     uint64 `json:"id"`
}

type Signtx struct {
	Rawtx       string          `json:"rawtx"`   //required	the raw transaction to extend
	Prevtxs     []SigntxPrevtxs `json:"prevtxs"` //required	a JSON array of transaction inputs ---->[]RawTxPrevtxs 数组json
	Privatekeys []string        `json:"privatekeys"`
}
type SigntxPrevtxs struct {
	Txid         string  `json:"txid"`                   // (string, required) the transaction hash
	Vout         int     `json:"vout"`                   // (number, required) the output number
	ScriptPubKey string  `json:"scriptPubKey"`           // (string, required) the output script
	Amount       float64 `json:"amount"`                 // (number, required) the output value
	RedeemScript string  `json:"redeemScript,omitempty"` //(string, required for P2SH or P2WSH) redeem script
}

//获取地址私钥
type ImportprivkeyResult struct {
	Result string `json:"result"` //返回结果
	Error  string `json:"error"`  //错误结果
	Id     uint64 `json:"id"`
}

type TxInput struct {
	Txins  []TxinPrevtxs `json:"txins"`
	Txouts []Txout       `json:"txouts"`
}

//未花费的余额
type TxinPrevtxs struct {
	Txid         string  `json:"txid"`                   //交易ID
	Vout         int     `json:"vout"`                   //vout位置
	ScriptPubKey string  `json:"scriptPubKey"`           //公钥
	Amount       float64 `json:"amount"`                 //当前utxo的金额
	RedeemScript string  `json:"redeemScript,omitempty"` //多签赎回脚本，一般单签为空即可
	Address      string  `json:"address"`                //本身构建交易不需要，但是签名的时候要去ab文件找对应的私钥签名
}

type Txout struct {
	ToAddress string  `json:"toAddress"`
	ToAmount  float64 `json:"toAmount"`
}

type RpcCreatetx struct {
	Txid string `json:"txid"` //交易ID
	Vout int    `json:"vout"` //vout位置
}

type AddressOutPut struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}
