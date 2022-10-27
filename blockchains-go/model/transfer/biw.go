package transfer

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

//Biw计价
type EstimateBiw struct {
	CoinType      string           `json:"coinType"`             //币种类型
	AppID         int              `json:"appId"`                //商户ID
	ChangeAddress string           `json:"changeAddress"`        //找零地址
	To            []EstimateOutBiw `json:"to"`                   //发送地址
	UseFee        int64            `json:"useFee"`               //手续费，可选项
	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutBiw struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type BiwUnspents struct {
	Code    int       `json:"code"`
	Data    []BiwUtxo `json:"data"`
	Message string    `json:"message"`
}

type BiwUtxo struct {
	Txid         string          `json:"txid"`
	Vout         int             `json:"vout"`
	Address      string          `json:"address"`
	AmountInt64  decimal.Decimal `json:"amount"`
	ScriptPubKey string          `json:"scriptPubKey"`
}

type BiwOrderRequest struct {
	OrderRequestHead
	Amount       int64                  `json:"amount,omitempty"`
	Fee          int64                  `json:"fee,omitempty"`
	OrderAddress []*BiwOrderAddrRequest `json:"order_address,omitempty"`
}

type BiwOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type BiwGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Biw验证地址====================
type BiwAddressResp struct {
	Code    int               `json:"code"`
	Data    *BiwAddressResult `json:"data"`
	Message string            `json:"message"`
}
type BiwAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Biw验证地址====================

//Biw unspents切片排序
type BiwUnspentDesc []BiwUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BiwUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BiwUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BiwUnspentDesc) Less(i, j int) bool {
	//return s[i].Amount > s[j].Amount
	return s[i].AmountInt64.GreaterThan(s[j].AmountInt64)
}

//Biw unspents切片排序
type BiwUnspentAsc []BiwUtxo

//实现排序三个接口
//为集合内元素的总数
func (s BiwUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BiwUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BiwUnspentAsc) Less(i, j int) bool {
	//return s[i].Amount < s[j].Amount
	return s[i].AmountInt64.LessThan(s[j].AmountInt64)
}

//========================json解析====================

func DecodeBiwGasResult(ds []byte) (*BiwGasResult, error) {
	ri := &BiwGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeBiwAddressResult(data []byte) *BiwAddressResp {
	if len(data) != 0 {
		result := new(BiwAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

//热钱包结构

type BiwTxTpl struct {
	MchId    string        `json:"mchId,omitempty"`
	OrderId  string        `json:"orderId,omitempty"`
	CoinName string        `json:"coinName"`
	TxIns    []BiwTxInTpl  `json:"txIns"`
	TxOuts   []BiwTxOutTpl `json:"txOuts"`
}

//utxo模板
type BiwTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type BiwTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
