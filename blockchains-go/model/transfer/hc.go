package transfer

import "github.com/shopspring/decimal"

type HcOrderRequest struct {
	OrderRequestHead
	FromAddress   string                  `json:"from_address"`   //非必须
	ChangeAddress string                  `json:"change_address"` //找零地址
	ToList        []*HcOrderToAddressList `json:"order_address"`
}

type HcRecycleOrderRequest struct {
	//Hash     string `json:"hash"`
	ApplyId    int64           `json:"applyId"`
	MchId      string          `json:"mchId"`
	OutOrderId string          `json:"orderOrderNo"`
	CoinName   string          `json:"coinName"`
	ToAddress  string          `json:"toAddress"`
	Model      int             `json:"model"`    //归集的时候使用 0 从小到大归集 1  从大到小归集
	FeeFloat   decimal.Decimal `json:"feeFloat"` //归集的时候使用 金额过小的时候需要指定一下手续费
}

type HcOrderToAddressList struct {
	Address  string `json:"address"` // '接收者地址'
	Quantity string `json:"quantity"`
}

type HcUnspents struct {
	Code    int      `json:"code"`
	Data    []HcUtxo `json:"data"`
	Message string   `json:"message"`
}

type HcUtxo struct {
	Txid          string `json:"txid"`
	Vout          int    `json:"vout"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
	ScriptPubKey  string `json:"scriptPubKey"`
}

//Hc unspents切片排序
type HcUnspentDesc []HcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s HcUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s HcUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s HcUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//Hc unspents切片排序
type HcUnspentAsc []HcUtxo

//实现排序三个接口
//为集合内元素的总数
func (s HcUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s HcUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s HcUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}
