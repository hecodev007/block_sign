package bo

import "github.com/group-coldwallet/zecserver/model"

//zec 签名模板

type ZecTxTpl struct {
	model.MchInfo
	ExpiryHeight int64         `json:"expiryHeight"`
	TxIns        []ZecTxInTpl  `json:"txIns"`
	TxOuts       []ZecTxOutTpl `json:"txOuts"`
	Hex          string        `json:"hex,omitempty"`
}

//utxo模板
type ZecTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromScriptPubKey string `json:"fromScriptPubKey"`           //来源UTXO的script
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}

//输出模板
type ZecTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
