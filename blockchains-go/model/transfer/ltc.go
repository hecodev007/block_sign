//package transfer
//
////Ltc计价
//type EstimateLtc struct {
//	CoinType      string           `json:"coinType"`      //币种类型
//	AppID         int              `json:"appId"`         //商户ID
//	ChangeAddress string           `json:"changeAddress"` //找零地址
//	To            []EstimateOutLtc `json:"to"`            //发送地址
//	UseFee        int64            `json:"useFee"`        //手续费，可选项
//}
//
//type EstimateOutLtc struct {
//	ToAddress string `json:"toAddr"`   //txout地址
//	ToAmount  int64  `json:"toAmount"` //txout金额
//}

package transfer

import "encoding/json"

//Ltc计价
type EstimateLtc struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutLtc `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutLtc struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type LtcUnspents struct {
	Code    int       `json:"code"`
	Data    []LtcUtxo `json:"data"`
	Message string    `json:"message"`
}

type LtcUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type LtcOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*LtcOrderAddrRequest `json:"order_address,omitempty"`
}

type LtcOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type LtcGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Ltc验证地址====================
type LtcAddressResp struct {
	Code    int               `json:"code"`
	Data    *LtcAddressResult `json:"data"`
	Message string            `json:"message"`
}
type LtcAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Ltc验证地址====================

//Ltc unspents切片排序
type LtcUnspentDesc []LtcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s LtcUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s LtcUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s LtcUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//Ltc unspents切片排序
type LtcUnspentAsc []LtcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s LtcUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s LtcUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s LtcUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeLtcGasResult(ds []byte) (*LtcGasResult, error) {
	ri := &LtcGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeLtcAddressResult(data []byte) *LtcAddressResp {
	if len(data) != 0 {
		result := new(LtcAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
