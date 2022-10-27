package transfer

import "encoding/json"

//BTC计价
type EstimateGhost struct {
	CoinType      string             `json:"coinType"`             //币种类型
	AppID         int                `json:"appId"`                //商户ID
	ChangeAddress string             `json:"changeAddress"`        //找零地址
	To            []EstimateOutGhost `json:"to"`                   //发送地址
	UseFee        int64              `json:"useFee"`               //手续费，可选项
	PayAddress    string             `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutGhost struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type GhostUnspents struct {
	Code    int         `json:"code"`
	Data    []GhostUtxo `json:"data"`
	Message string      `json:"message"`
}

type GhostUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

type GhostOrderRequest struct {
	OrderRequestHead
	ChangeAddr string         `json:"changeAddr"`
	Fee        int64          `json:"fee"`
	TxIns      []*GhostTxins  `json:"txIns"`
	TxOut      []*GhostTxOuts `json:"txOuts"`
}

type GhostTxins struct {
	FromAddr   string `json:"fromAddr"`
	FromTxId   string `json:"fromTxid"`
	FromIndex  int    `json:"fromIndex"`
	FromAmount int64  `json:"fromAmount"`
}

type GhostTxOuts struct {
	ToAddr   string `json:"toAddr"`
	ToAmount int64  `json:"toAmount"`
}

//====================手续费请求结果====================
type GhostGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================btc验证地址====================
type GhostAddressResp struct {
	Code    int                 `json:"code"`
	Data    *GhostAddressResult `json:"data"`
	Message string              `json:"message"`
}
type GhostAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================btc验证地址====================

//BTC unspents切片排序
type GhostUnspentDesc []GhostUtxo

//实现排序三个接口
//为集合内元素的总数
func (s GhostUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s GhostUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s GhostUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//BTC unspents切片排序
type GhostUnspentAsc []GhostUtxo

//实现排序三个接口
//为集合内元素的总数
func (s GhostUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s GhostUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s GhostUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

//========================json解析====================

func DecodeGhostGasResult(ds []byte) (*GhostGasResult, error) {
	ri := &GhostGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeGhostAddressResult(data []byte) *GhostAddressResp {
	if len(data) != 0 {
		result := new(GhostAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
