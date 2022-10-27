package rylink

import (
	"fmt"
	"github.com/group-coldwallet/bchsign/model/bo"
	"testing"
)

func TestBchSignTxTpl(t *testing.T) {
	tpl := &bo.BchTxTpl{
		TxIns: []bo.BchTxInTpl{(bo.BchTxInTpl{
			FromAddr:    "qrxny22vj967nygks5fgm33r6984cl7zlu9h5madq3",
			FromPrivkey: "L35bbBscBwYp7VhWeYRnFsLRZvRARDo9ot45H3jtp3vty36cH7FJ",
			FromTxid:    "65daecb3a56b08f48e5e6aa37a1e2d7f741e494bf6f6034e264551018fa3b2a6",
			FromIndex:   1,
			FromAmount:  150328,
		})},
		TxOuts: []bo.BchTxOutTpl{bo.BchTxOutTpl{
			ToAddr:   "qrxny22vj967nygks5fgm33r6984cl7zlu9h5madq3",
			ToAmount: 147690,
		}, bo.BchTxOutTpl{
			ToAddr:   "1KhycbX9fvG81anH5WFqGcWfu9mQvFxZpV",
			ToAmount: 546,
		}, bo.BchTxOutTpl{
			ToAddr:   "pqk0s7uywvj48p2z89z4zhzh0gg33eef8uw6ymv638",
			ToAmount: 546,
		}, bo.BchTxOutTpl{
			ToAddr:   "35noKrGr9sz5zL2cBE1ZJ6PsDnpRYLp8DB",
			ToAmount: 546,
		}},
	}
	fmt.Println(BchSignTxTpl(tpl))

}

func TestDecodeAddr(t *testing.T) {
	//_, _, err := CreatePayScript("3HUmB9y9QhroSyHu3uXpJZipQbCjh2UBpL")
	_, _, err := CreatePayScript("3FsKNWwRxmM2YRQpZtRZ9zGhnk2rt5ZHHf")
	if err != nil {
		t.Log(err.Error())
	} else {
		t.Log("pass")
	}

}
