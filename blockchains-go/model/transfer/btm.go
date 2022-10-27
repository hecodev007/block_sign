package transfer

import (
	"encoding/json"
	"fmt"
)

type BtmOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	//  token 转账 需要参数
	ContractAddress string `json:"contract_address"`
	Token           string `json:"token"` // 代币的名字，主链转账不传这个值
}

func DecodeBtmTransferResp(data []byte) map[string]interface{} {
	var result map[string]interface{}
	if len(data) != 0 {
		err := json.Unmarshal(data, &result)
		if err == nil {
			return result
		} else {
			fmt.Printf("parse response data error,err=%v", err)
		}
	}
	return nil
}

//package transfer
//
//import (
//	"encoding/json"
//	"github.com/shopspring/decimal"
//)
//
////Sat计价
//type EstimateBtm struct {
//	CoinType      string           `json:"coinType"`             //币种类型
//	AppID         int              `json:"appId"`                //商户ID
//	ChangeAddress string           `json:"changeAddress"`        //找零地址
//	To            []EstimateOutBtm `json:"to"`                   //发送地址
//	UseFee        int64            `json:"useFee"`               //手续费，可选项
//	PayAddress    string           `json:"payAddress,omitempty"` //目前usdt的可选代付地址
//}
//
//type EstimateOutBtm struct {
//	ToAddress     string `json:"toAddr"`   //txout地址
//	ToAmount      int64  `json:"toAmount"` //txout金额
//	ToTokenAmonut int64  `json:"toTokenAmonut,omitempty"`
//}
//
//type BtmUnspents struct {
//	Code    int       `json:"code"`
//	Data    []BtmUtxo `json:"data"`
//	Message string    `json:"message"`
//}
//
//type BtmUtxo struct {
//	Txid         string          `json:"txid"`
//	Vout         int             `json:"vout"`
//	Address      string          `json:"address"`
//	AmountInt64  decimal.Decimal `json:"amount"`
//	ScriptPubKey string          `json:"scriptPubKey"`
//}
//
//type BtmOrderRequest struct {
//	OrderRequestHead
//	Amount       int64                  `json:"amount,omitempty"`
//	Fee          int64                  `json:"fee,omitempty"`
//	OrderAddress []*BtmOrderAddrRequest `json:"order_address,omitempty"`
//}
//
//type BtmOrderAddrRequest struct {
//	Dir     DirType `json:"dir"`
//	Address string  `json:"address"`
//	Amount  int64   `json:"amount"`
//	TxID    string  `json:"txId"`
//	Vout    int     `json:"vout"`
//}
//
////====================手续费请求结果====================
//type BtmGasResult struct {
//	FastestFee  int64
//	HalfHourFee int64
//	HourFee     int64
//}
//
////====================手续费请求结果====================
//
////====================Btm验证地址====================
//type BtmAddressResp struct {
//	Code    int               `json:"code"`
//	Data    *BtmAddressResult `json:"data"`
//	Message string            `json:"message"`
//}
//type BtmAddressResult struct {
//	Isvalid bool `json:"isvalid"`
//}
//
////====================Btm验证地址====================
//
////Btm unspents切片排序
//type BtmUnspentDesc []BtmUtxo
//
////实现排序三个接口
////为集合内元素的总数
//func (s BtmUnspentDesc) Len() int {
//	return len(s)
//}
//
////Swap 交换索引为 i 和 j 的元素
//func (s BtmUnspentDesc) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//
////从大到小，最大金额排序
////如果index为i的元素大于index为j的元素，则返回true，否则返回false
//func (s BtmUnspentDesc) Less(i, j int) bool {
//	//return s[i].Amount > s[j].Amount
//	return s[i].AmountInt64.GreaterThan(s[j].AmountInt64)
//}
//
////Btm unspents切片排序
//type BtmUnspentAsc []BtmUtxo
//
////实现排序三个接口
////为集合内元素的总数
//func (s BtmUnspentAsc) Len() int {
//	return len(s)
//}
//
////Swap 交换索引为 i 和 j 的元素
//func (s BtmUnspentAsc) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//
////从小到大，最小金额排序
////如果index为i的元素大于index为j的元素，则返回true，否则返回false
//func (s BtmUnspentAsc) Less(i, j int) bool {
//	//return s[i].Amount < s[j].Amount
//	return s[i].AmountInt64.LessThan(s[j].AmountInt64)
//}
//
////========================json解析====================
//
//func DecodeBtmGasResult(ds []byte) (*BtmGasResult, error) {
//	ri := &BtmGasResult{}
//	err := json.Unmarshal(ds, &ri)
//	if err != nil {
//		return nil, err
//	}
//	return ri, nil
//}
//
//func DecodeBtmAddressResult(data []byte) *BtmAddressResp {
//	if len(data) != 0 {
//		result := new(BtmAddressResp)
//		err := json.Unmarshal(data, result)
//		if err == nil {
//			return result
//		}
//	}
//	return nil
//}
//
////热钱包结构
//
//type BtmTxTpl struct {
//	MchId    string        `json:"mchId,omitempty"`
//	OrderId  string        `json:"orderId,omitempty"`
//	CoinName string        `json:"coinName"`
//	TxIns    []BtmTxInTpl  `json:"txIns"`
//	TxOuts   []BtmTxOutTpl `json:"txOuts"`
//}
//
////utxo模板
//type BtmTxInTpl struct {
//	FromAddr         string `json:"fromAddr"`                   //来源地址
//	FromPrivkey      string `json:"fromPrivkey,omitempty"`      //来源地址地址对于的私钥，签名期间赋值
//	FromTxid         string `json:"fromTxid"`                   //来源UTXO的txid
//	FromIndex        uint32 `json:"fromIndex"`                  //来源UTXO的txid 地址的下标
//	FromAmount       int64  `json:"fromAmount"`                 //来源UTXO的txid 对应的金额
//	FromRedeemScript string `json:"fromRedeemScript,omitempty"` //多签脚本
//
//}
//
////输出模板
//type BtmTxOutTpl struct {
//	ToAddr   string `json:"toAddr"`   //txout地址
//	ToAmount int64  `json:"toAmount"` //txout金额
//}
