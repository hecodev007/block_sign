package bo

import "github.com/group-coldwallet/ucaserver/model"

//uca 签名模板

type UcaTxTpl struct {
	model.MchInfo
	TxIns  []UcaTxInTpl  `json:"txIns"` //如果是
	TxOuts []UcaTxOutTpl `json:"txOuts"`
}

//utxo模板
type UcaTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type UcaTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
