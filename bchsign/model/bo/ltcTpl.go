package bo

import "github.com/group-coldwallet/bchsign/model"

//bch 签名模板

type BchTxTpl struct {
	model.MchInfo
	TxIns  []BchTxInTpl  `json:"txIns"`
	TxOuts []BchTxOutTpl `json:"txOuts"`
}

//utxo模板
type BchTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}

//输出模板
type BchTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
