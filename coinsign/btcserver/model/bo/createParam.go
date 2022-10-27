package bo

import "github.com/group-coldwallet/btcserver/model"

type TxInput struct {
	Txins      []TxinPrevtxs `json:"txins"`
	Txouts     []Txout       `json:"txouts"`
	ChangeAddr string        `json:"changeAddr"`
	Fee        int64         `json:"fee"`
	model.MchInfo
}

//未花费的余额
type TxinPrevtxs struct {
	Txid         string `json:"txid"`                   //交易ID
	Vout         int    `json:"vout"`                   //vout位置
	ScriptPubKey string `json:"scriptPubKey"`           //公钥
	Amount       int64  `json:"amount"`                 //当前utxo的金额
	RedeemScript string `json:"redeemScript,omitempty"` //多签赎回脚本，一般单签为空即可
	Address      string `json:"address"`                //本身构建交易不需要，但是签名的时候要去ab文件找对应的私钥签名
}

type Txout struct {
	ToAddress string `json:"toAddress"`
	ToAmount  int64  `json:"toAmount"`
}

type RpcCreatetx struct {
	Txid string `json:"txid"` //交易ID
	Vout int    `json:"vout"` //vout位置
}
