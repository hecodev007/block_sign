package neo_model

import "github.com/shopspring/decimal"

//
//{
//"jsonrpc": "2.0",
//"id": 1,
//"result": {
//"txid": "0x80770d066e6764772f303608d18f5820c2bd45fc8bb4909ddc36a8f55284c9e4",
//"size": 202,
//"type": "ContractTransaction",
//"version": 0,
//"attributes": [],
//"vin": [
//{
//"txid": "0x7a4acf96c83065c12e18b7db6fdc2ffbed2262623aeed6f5e2b807d57eccef37",
//"vout": 0
//}
//],
//"vout": [
//{
//"n": 0,
//"asset": "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
//"value": "1",
//"address": "AT9RuZS2if6dKSf6T7wCgfqpon8KQatZp2"
//}
//],
//"sys_fee": "0",
//"net_fee": "0",
//"scripts": [
//{
//"invocation": "40e8d8e71d8c4b34545bcd5718176d45d001a77bb822bc12a025a05de31e386d79d2ed853e8095dccf183e44d4829eae25c1665fc983748e0276d8a4fef3e3aff6",
//"verification": "2102195db806b08074711694c7902e604419706c4d372113edb27b453007a3c474c3ac"
//}
//],
//"blockhash": "0x12cb987b422e0b73a1167e223afa4f30234a098293e8394be95460e253ac109e",
//"confirmations": 545,
//"blocktime": 1605695868
//}
//}

type GetrawtransactionResp struct {
	Result *GetrawtransactionResult `json:"result"`
}

type GetrawtransactionResult struct {
	Txid          string `json:"txid"`
	Type          string `json:"type"` //只是判断主链币交易 ContractTransaction 类型即可
	Attributes    []*Attributes
	Vin           []*Vin
	Vout          []*Vout
	SysFee        decimal.Decimal `json:"sys_fee"` //目前一直都是0
	NetFee        decimal.Decimal `json:"net_fee"` //目前一直都是0
	Blockhash     string          `json:"blockhash"`
	Confirmations int64           `json:"confirmations"`
	Blocktime     int64           `json:"blocktime"`
	//Scripts interface{} //暂时先不解析脚本

}

type Attributes struct {
	Usage string `json:"usage"` //通常为Script
	Data  string `json:"data"`
}

type Vin struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

type Vout struct {
	N       int             `json:"n"`
	Asset   string          `json:"asset"` //资产ID ，只需要识别主链币即可 //0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b
	Value   decimal.Decimal `json:"value"`
	Address string          `json:"address"`
}

type GetblockResp struct {
	Result struct {
		Index int64 `json:"index"` //高度
	}
}
