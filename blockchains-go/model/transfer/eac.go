package transfer

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

//Eac计价
type EstimateEac struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutEac `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutEac struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type EacUnspents struct {
	Code    int       `json:"code"`
	Data    []EacUtxo `json:"data"`
	Message string    `json:"message"`
}

type EacUtxo struct {
	Txid         string          `json:"txid"`
	Vout         int             `json:"vout"`
	Address      string          `json:"address"`
	AmountInt64  decimal.Decimal `json:"amount"`
	ScriptPubKey string          `json:"scriptPubKey"`
}

type EacOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*EacOrderAddrRequest `json:"order_address,omitempty"`
}

type EacOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type EacGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Eac验证地址====================
type EacAddressResp struct {
	Code    int               `json:"code"`
	Data    *EacAddressResult `json:"data"`
	Message string            `json:"message"`
}
type EacAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Eac验证地址====================

//Eac unspents切片排序
type EacUnspentDesc []EacUtxo

//实现排序三个接口
//为集合内元素的总数
func (s EacUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s EacUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s EacUnspentDesc) Less(i, j int) bool {
	//return s[i].Amount > s[j].Amount
	return s[i].AmountInt64.GreaterThan(s[j].AmountInt64)
}

//Eac unspents切片排序
type EacUnspentAsc []EacUtxo

//实现排序三个接口
//为集合内元素的总数
func (s EacUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s EacUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s EacUnspentAsc) Less(i, j int) bool {
	//return s[i].Amount < s[j].Amount
	return s[i].AmountInt64.LessThan(s[j].AmountInt64)
}

//========================json解析====================

func DecodeEacGasResult(ds []byte) (*EacGasResult, error) {
	ri := &EacGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeEacAddressResult(data []byte) *EacAddressResp {
	if len(data) != 0 {
		result := new(EacAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

//热钱包结构

type EacTxTpl struct {
	MchId    string        `json:"mchId,omitempty"`
	OrderId  string        `json:"orderId,omitempty"`
	CoinName string        `json:"coinName"`
	TxIns    []EacTxInTpl  `json:"txIns"`
	TxOuts   []EacTxOutTpl `json:"txOuts"`
}

//utxo模板
type EacTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type EacTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
