package transfer

import (
	"encoding/json"
)

//Zen计价
type EstimateZen struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutZen `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutZen struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type ZenUnspents struct {
	Code    int       `json:"code"`
	Data    []ZenUtxo `json:"data"`
	Message string    `json:"message"`
}

type ZenUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type ZenOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*ZenOrderAddrRequest `json:"order_address,omitempty"`
}

type ZenOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type ZenGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Zen验证地址====================
type ZenAddressResp struct {
	Code    int               `json:"code"`
	Data    *ZenAddressResult `json:"data"`
	Message string            `json:"message"`
}
type ZenAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Zen验证地址====================

//Zen unspents切片排序
type ZenUnspentDesc []ZenUtxo

//实现排序三个接口
//为集合内元素的总数
func (s ZenUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s ZenUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s ZenUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//Zen unspents切片排序
type ZenUnspentAsc []ZenUtxo

//实现排序三个接口
//为集合内元素的总数
func (s ZenUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s ZenUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s ZenUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeZenGasResult(ds []byte) (*ZenGasResult, error) {
	ri := &ZenGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeZenAddressResult(data []byte) *ZenAddressResp {
	if len(data) != 0 {
		result := new(ZenAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

//热钱包结构

type ZenTxTpl struct {
	MchId    string        `json:"mchId,omitempty"`
	OrderId  string        `json:"orderId,omitempty"`
	CoinName string        `json:"coinName"`
	TxIns    []ZenTxInTpl  `json:"txIns"`
	TxOuts   []ZenTxOutTpl `json:"txOuts"`
}

//utxo模板
type ZenTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type ZenTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
