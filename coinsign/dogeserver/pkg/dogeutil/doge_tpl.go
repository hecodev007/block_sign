package dogeutil

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

//doge 签名模板

type DogeTxTpl struct {
	MchId    string         `json:"mchId,omitempty"`
	OrderId  string         `json:"orderId,omitempty"`
	CoinName string         `json:"coinName,omitempty"`
	TxIns    []DogeTxInTpl  `json:"txIns"` //如果是
	TxOuts   []DogeTxOutTpl `json:"txOuts"`
}

//utxo模板
type DogeTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmountInt64  int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
}

//输出模板
type DogeTxOutTpl struct {
	ToAddr        string `json:"toAddr"`   //txout地址
	ToAmountInt64 int64  `json:"toAmount"` //txout金额
}

func (tpl *DogeTxTpl) Check() error {
	fromFloatTotal := decimal.Zero
	toFloatTotal := decimal.Zero
	for _, v := range tpl.TxIns {
		am := decimal.NewFromInt(v.FromAmountInt64)
		if am.IsZero() {
			return fmt.Errorf("error amount,address:%s", v.FromAddr)
		}
		if v.FromAddr == "" || v.FromTxid == "" || v.FromPrivkey == "" {
			return errors.New("error params")
		}
		fromFloatTotal = fromFloatTotal.Add(am)
	}

	for _, v := range tpl.TxOuts {
		am := decimal.NewFromInt(v.ToAmountInt64)
		if am.IsZero() {
			return fmt.Errorf("error amount,to address,%s", am.String())
		}
		if v.ToAddr == "" {
			return errors.New("error out params")
		}
		toFloatTotal = toFloatTotal.Add(am)
	}
	//手续费不允许超过50个币
	if fromFloatTotal.Sub(toFloatTotal).GreaterThan(decimal.NewFromInt(50)) {
		return errors.New("fee, too hight")
	}
	return nil
}
