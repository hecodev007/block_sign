package ucautil

import (
	"github.com/group-coldwallet/ucaserver/model/bo"
	"github.com/shopspring/decimal"
	"testing"
)

//
//{
//"result": [
//{
//"txid": "03cfa9cbd9b8b0fb0e9801bb85edce8004d454ce74ee7083d88d148d53a838f9",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 0.03330000,
//"confirmations": 3,
//"spendable": true
//},
//{
//"txid": "0462a5d752fb4ff588197f24481e1e83bc1c0f51dabaa8ce551dfef39b3f3b68",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 1.99999999,
//"confirmations": 68,
//"spendable": true
//},
//{
//"txid": "12373b66e9e09334ba9f9e651dc6c425222e423aaf363422486a3830cdfbe4e1",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 1.90000000,
//"confirmations": 49,
//"spendable": true
//},
//{
//"txid": "284f5655eb2a80ab11eb3975be7ba4d9ac2f83e4b56b2286de0102e388ecde82",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 2.10000000,
//"confirmations": 145,
//"spendable": true
//},
//{
//"txid": "509f0c0ddd04783b2cd42407b5e8053e90a3fc63c2b1bd9232e5b7c23ef05a45",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 3.19999990,
//"confirmations": 148,
//"spendable": true
//},
//{
//"txid": "619a1af7b6843fe8f41aaffcbfe30db23436f82575ed9bbce7f7ee2f422f5294",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 0.04123000,
//"confirmations": 36,
//"spendable": true
//},
//{
//"txid": "76697efff3b44df0815bb2705d7b0335a4fb4ee8158f8452aa9e378fe2586c01",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 0.19999900,
//"confirmations": 168,
//"spendable": true
//},
//{
//"txid": "7f0c669a180ef3caabf8e35584cbdc50f1d903d98b2143404c8e781bf714ae49",
//"vout": 0,
//"address": "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
//"account": "",
//"scriptPubKey": "76a914a09b815b0ea45b6351a991d5e6d9146951cc328788ac",
//"amount": 0.05000000,
//"confirmations": 43,
//"spendable": true
//}
//],
//"error": null,
//"id": 11
//}

func TestSignTxTpl(t *testing.T) {
	tpl := &bo.UcaTxTpl{
		TxIns: []bo.UcaTxInTpl{
			bo.UcaTxInTpl{
				FromAddr:    "UcdEMmuKkH5zi1cakX1A4bCMeRCyCEmk19",
				FromPrivkey: "VT52WZEzwbfUZUrGu114PtZP3fKxDydb4wzhAQ61x1JwyFAqT1ik",
				FromTxid:    "45f252c252dc2ee607df597572a037219cbef9aeecae41051bce5a4ca6a8bf3d",
				FromIndex:   uint32(1),
				FromAmount:  decimal.NewFromFloat(4.52152889).Shift(8).IntPart(),
			},
		},
		TxOuts: []bo.UcaTxOutTpl{
			bo.UcaTxOutTpl{
				ToAddr:   "UkYJQ8gUaEz8iQC3Kc11R9c6QHDnvfNZnM",
				ToAmount: int64(452052889),
			},
		},
	}
	t.Log(SignTxTpl(tpl))
}
