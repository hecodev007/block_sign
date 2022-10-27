//package transfer
//
////Bch计价
//type EstimateBch struct {
//	CoinType      string           `json:"coinType"`      //币种类型
//	AppID         int              `json:"appId"`         //商户ID
//	ChangeAddress string           `json:"changeAddress"` //找零地址
//	To            []EstimateOutBch `json:"to"`            //发送地址
//	UseFee        int64            `json:"useFee"`        //手续费，可选项
//}
//
//type EstimateOutBch struct {
//	ToAddress string `json:"toAddr"`   //txout地址
//	ToAmount  int64  `json:"toAmount"` //txout金额
//}

package transfer

import "encoding/json"

//Bch计价
type EstimateBch struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutBch `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutBch struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type BchUnspents struct {
	Code    int       `json:"code"`
	Data    []BchUtxo `json:"data"`
	Message string    `json:"message"`
}

type BchUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type BchOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*BchOrderAddrRequest `json:"order_address,omitempty"`
}

type BchOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type BchGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Bch验证地址====================
type BchAddressResp struct {
	Code    int               `json:"code"`
	Data    *BchAddressResult `json:"data"`
	Message string            `json:"message"`
}
type BchAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Bch验证地址====================

//Bch unspents切片排序
type BchUnspentDesc []BchUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BchUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BchUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BchUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//Bch unspents切片排序
type BchUnspentAsc []BchUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BchUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BchUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BchUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeBchGasResult(ds []byte) (*BchGasResult, error) {
	ri := &BchGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeBchAddressResult(data []byte) *BchAddressResp {
	if len(data) != 0 {
		result := new(BchAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
