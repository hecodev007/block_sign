package transfer

import "encoding/json"

type DcrOrderRequest struct {
	OrderRequestHead
	Data *DcrSignReq `json:"data"`
}
type DcrSignReq struct {
	RawTx     string        `json:"raw_tx"`
	Addresses []string      `json:"addresses"`
	Inputs    []*RawTxInput `json:"inputs"`
}

type RawTxInput struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	Tree         int8   `json:"tree"`
	ScriptPubKey string `json:"scriptPubKey"`
	RedeemScript string `json:"redeemScript"`
}

type DcrTxTpl struct {
	MchId    string        `json:"mchId,omitempty"`
	OrderId  string        `json:"orderId,omitempty"`
	CoinName string        `json:"coinName"`
	TxIns    []DcrTxInTpl  `json:"txIns"`
	TxOuts   []DcrTxOutTpl `json:"txOuts"`
}

//utxo模板
type DcrTxInTpl struct {
	FromAddr         string  `json:"fromAddr"`                   //来源地址
	FromPrivkey      string  `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string  `json:"txid"`                       //来源UTXO的txid			//modify by flynn
	FromIndex        uint32  `json:"vout"`                       //来源UTXO的txid 地址的下标 //modify by flynn
	FromAmount       float64 `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string  `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type DcrTxOutTpl struct {
	ToAddr   string  `json:"toAddr"`   //txout地址
	ToAmount float64 `json:"toAmount"` //txout金额
}

type DcrCreateTxReq struct {
	Vin  []DcrTxInTpl  `json:"vin"`
	Vout []DcrTxOutTpl `json:"vout"`
}

type DcrAddressResp struct {
	Code    int               `json:"code"`
	Data    *DcrAddressResult `json:"data"`
	Message string            `json:"message"`
}
type DcrAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

func DecodeDcrAddressResult(data []byte) *DcrAddressResp {
	if len(data) != 0 {
		result := new(DcrAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
