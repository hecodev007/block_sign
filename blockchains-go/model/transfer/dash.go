package transfer

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

//Dash计价
type EstimateDash struct {
	CoinType      string            `json:"coinType"`             //币种类型
	AppID         int               `json:"appId"`                //商户ID
	ChangeAddress string            `json:"changeAddress"`        //找零地址
	To            []EstimateOutDash `json:"to"`                   //发送地址
	UseFee        int64             `json:"useFee"`               //手续费，可选项
	PayAddress    string            `json:"payAddress,omitempty"` //目前usdt的可选代付地址
}

type EstimateOutDash struct {
	ToAddress     string `json:"toAddr"`   //txout地址
	ToAmount      int64  `json:"toAmount"` //txout金额
	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
}

type DashUnspents struct {
	Code    int        `json:"code"`
	Data    []DashUtxo `json:"data"`
	Message string     `json:"message"`
}

type DashUtxo struct {
	Txid         string          `json:"txid"`
	Vout         int             `json:"outputIndex"`
	Address      string          `json:"address"`
	AmountFloat  decimal.Decimal `json:"satoshis"`
	ScriptPubKey string          `json:"script"`
}

type DashOrderRequest struct {
	OrderRequestHead
	Amount       int64                   `json:"amount,omitempty"`
	Fee          int64                   `json:"fee,omitempty"`
	OrderAddress []*DashOrderAddrRequest `json:"order_address,omitempty"`
}

type DashOrderAddrRequest struct {
	Dir     DirType `json:"dir"`
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	TxID    string  `json:"txId"`
	Vout    int     `json:"vout"`
}

//====================手续费请求结果====================
type DashGasResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//====================手续费请求结果====================

//====================Dash验证地址====================
type DashAddressResp struct {
	Code    int                `json:"code"`
	Data    *DashAddressResult `json:"data"`
	Message string             `json:"message"`
}
type DashAddressResult struct {
	Isvalid bool `json:"isvalid"`
}

//====================Dash验证地址====================

//Dash unspents切片排序
type DashUnspentDesc []DashUtxo

//实现排序三个接口
//为集合内元素的总数
func (s DashUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s DashUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s DashUnspentDesc) Less(i, j int) bool {
	//return s[i].Amount > s[j].Amount
	return s[i].AmountFloat.GreaterThan(s[j].AmountFloat)
}

//Dash unspents切片排序
type DashUnspentAsc []DashUtxo

//实现排序三个接口
//为集合内元素的总数
func (s DashUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s DashUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s DashUnspentAsc) Less(i, j int) bool {
	//return s[i].Amount < s[j].Amount
	return s[i].AmountFloat.LessThan(s[j].AmountFloat)
}

//========================json解析====================

func DecodeDashGasResult(ds []byte) (*DashGasResult, error) {
	ri := &DashGasResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeDashAddressResult(data []byte) *DashAddressResp {
	if len(data) != 0 {
		result := new(DashAddressResp)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

//热钱包结构

type DashTxTpl struct {
	MchId    string         `json:"mchId,omitempty"`
	OrderId  string         `json:"orderId,omitempty"`
	CoinName string         `json:"coinName"`
	TxIns    []DashTxInTpl  `json:"txIns"`
	TxOuts   []DashTxOutTpl `json:"txOuts"`
}

//utxo模板
type DashTxInTpl struct {
	FromAddr         string `json:"fromAddr"`                   //来源地址
	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本

}

//输出模板
type DashTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
