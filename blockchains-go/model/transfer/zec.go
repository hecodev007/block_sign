//package transfer
//
////Zec计价
//type EstimateZec struct {
//	CoinType      string           `json:"coinType"`      //币种类型
//	AppID         int              `json:"appId"`         //商户ID
//	ChangeAddress string           `json:"changeAddress"` //找零地址
//	To            []EstimateOutZec `json:"to"`            //发送地址
//	UseFee        int64            `json:"useFee"`        //手续费，可选项
//}
//
//type EstimateOutZec struct {
//	ToAddress string `json:"toAddr"`   //txout地址
//	ToAmount  int64  `json:"toAmount"` //txout金额
//}

package transfer

import "encoding/json"

//Zec计价
type EstimateZec struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutZec `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutZec struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type ZecUnspents struct {
	Code    int       `json:"code"`
	Data    []ZecUtxo `json:"data"`
	Message string    `json:"message"`
}

type ZecUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type ZecOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	ExpiryHeight int64                  `json:"expiryHeight"`
	OrderAddress []*ZecOrderAddrRequest `json:"order_address,omitempty"`
}

type ZecOrderAddrRequest struct {
	Dir          DirType `json:"dir"`
	Address      string  `json:"address"`
	Amount       int64   `json:"amount"`
	TxID         string  `json:"txId"`
	Vout         int     `json:"vout"`
	ScriptPubKey string  `json:"scriptPubKey"`
}

//====================手续费请求结果====================
type ZecGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Zec验证地址====================
type ZecAddressResp struct {
	Code    int               `json:"code"`
	Data    *ZecAddressResult `json:"data"`
	Message string            `json:"message"`
}
type ZecAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Zec验证地址====================

//Zec unspents切片排序
type ZecUnspentDesc []ZecUtxo

//实现排序三个接口
//为集合内元素的总数
func (s ZecUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s ZecUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s ZecUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//Zec unspents切片排序
type ZecUnspentAsc []ZecUtxo

//实现排序三个接口
//为集合内元素的总数
func (s ZecUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s ZecUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s ZecUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeZecGasResult(ds []byte) (*ZecGasResult, error) {
	ri := &ZecGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeZecAddressResult(data []byte) *ZecAddressResp {
	if len(data) != 0 {
		result := new(ZecAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

type ZecHeightResp struct {
	Code    int              `json:"code"`
	Data    *ZecHeightResult `json:"data"`
	Message string           `json:"message"`
}

type ZecHeightResult struct {
	Headers int64 `json:"headers"`
}

func DecodeZecHeightResult(data []byte) *ZecHeightResp {
	if len(data) != 0 {
		result := new(ZecHeightResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
