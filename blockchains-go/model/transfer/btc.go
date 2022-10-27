package transfer

import "encoding/json"

//BTC计价
type EstimateBtc struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutBtc `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutBtc struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type BtcUnspents struct {
	Code    int       `json:"code"`
	Data    []BtcUtxo `json:"data"`
	Message string    `json:"message"`
}

type BtcUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type BtcOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*BtcOrderAddrRequest `json:"order_address,omitempty"`
}

type BtcOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type BtcGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================btc验证地址====================
type BtcAddressResp struct {
	Code    int               `json:"code"`
	Data    *BtcAddressResult `json:"data"`
	Message string            `json:"message"`
}
type BtcAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================btc验证地址====================

//BTC unspents切片排序
type BtcUnspentDesc []BtcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BtcUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BtcUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BtcUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//BTC unspents切片排序
type BtcUnspentAsc []BtcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BtcUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BtcUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BtcUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeBtcGasResult(ds []byte) (*BtcGasResult, error) {
	ri := &BtcGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeBtcAddressResult(data []byte) *BtcAddressResp {
	if len(data) != 0 {
		result := new(BtcAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
