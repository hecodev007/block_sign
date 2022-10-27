package transfer

import (
	"encoding/json"
)

//Ada计价
type EstimateAda struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutAda `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutAda struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type AdaUnspents struct {
	Code    int       `json:"code"`
	Data    []AdaUtxo `json:"data"`
	Message string    `json:"message"`
}

type AdaUtxo struct {
	Txid    string           `json:"txid"`
	Vout    int              `json:"vout"`
	Address string           `json:"address"`
	Tokens  map[string]int64 `json:"tokens"`
}

type AdaOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*AdaOrderAddrRequest `json:"order_address,omitempty"`
}

type AdaOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type AdaGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Ada验证地址====================
type AdaAddressResp struct {
	Code    int               `json:"code"`
	Data    *AdaAddressResult `json:"data"`
	Message string            `json:"message"`
}
type AdaAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Ada验证地址====================

//Ada unspents切片排序
type AdaUnspentDesc []AdaUtxo

//实现排序三个接口
//为集合内元素的总数
func (s AdaUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s AdaUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AdaUnspentDesc) Less(i, j int) bool {
	return s[i].Tokens["ADA"] > s[j].Tokens["ADA"]
}

//Ada unspents切片排序
type AdaTokenUnspentDesc struct {
	Assertid string
	Utxos    []AdaUtxo
}

//实现排序三个接口
//为集合内元素的总数
func (s AdaTokenUnspentDesc) Len() int {
	return len(s.Utxos)
}

//Swap 交换索引为 i 和 j 的元素
func (s AdaTokenUnspentDesc) Swap(i, j int) {
	s.Utxos[i], s.Utxos[j] = s.Utxos[j], s.Utxos[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AdaTokenUnspentDesc) Less(i, j int) bool {
	return s.Utxos[i].Tokens[s.Assertid] > s.Utxos[j].Tokens[s.Assertid]
}

//Ada unspents切片排序
type AdaUnspentAsc []AdaUtxo

//实现排序三个接口
//为集合内元素的总数
func (s AdaUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s AdaUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AdaUnspentAsc) Less(i, j int) bool {
	return s[i].Tokens["ADA"] < s[j].Tokens["ADA"]
}

type AdaTokenUnspentAsc struct {
	Assertid string
	Utxos    []AdaUtxo
}

func (s AdaTokenUnspentAsc) Len() int {
	return len(s.Utxos)
}

//Swap 交换索引为 i 和 j 的元素
func (s AdaTokenUnspentAsc) Swap(i, j int) {
	s.Utxos[i], s.Utxos[j] = s.Utxos[j], s.Utxos[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AdaTokenUnspentAsc) Less(i, j int) bool {
	return s.Utxos[i].Tokens[s.Assertid] < s.Utxos[j].Tokens[s.Assertid]
}

//========================json解析====================

func DecodeAdaGasResult(ds []byte) (*AdaGasResult, error) {
	ri := &AdaGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeAdaAddressResult(data []byte) *AdaAddressResp {
	if len(data) != 0 {
		result := new(AdaAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

//热钱包结构

type AdaTxTpl struct {
	MchId    string        `json:"mchId,omitempty"`
	OrderId  string        `json:"orderId,omitempty"`
	CoinName string        `json:"coinName"`
	TxIns    []AdaTxInTpl  `json:"txIns"`
	TxOuts   []AdaTxOutTpl `json:"txOuts"`
	Change   string        `json:"change"`
}

//utxo模板
type AdaTxInTpl struct {
	FromAddr string `json:"fromAddr"` //来源地址
	//FromPrivkey string           `json:"fromPrivkey,omitempty"` //来源地址地址对于的私钥，签名期间赋值
	FromTxid  string           `json:"fromTxid"`  //来源UTXO的txid
	FromIndex uint32           `json:"fromIndex"` //来源UTXO的txid 地址的下标
	Tokens    map[string]int64 `json:"tokens"`
}

//输出模板
type AdaTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
	Token    string `json:"token"`
}
