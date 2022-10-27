package rylink

import (
	"fmt"
	"github.com/group-coldwallet/bchserver/model/bo"
	"testing"
)

func TestBchSignTxTpl(t *testing.T) {
	tpl := &bo.BchTxTpl{
		TxIns: []bo.BchTxInTpl{(bo.BchTxInTpl{
			FromAddr:    "qrxny22vj967nygks5fgm33r6984cl7zlu9h5madq3",
			FromPrivkey: "L35bbBscBwYp7VhWeYRnFsLRZvRARDo9ot45H3jtp3vty36cH7FJ",
			FromTxid:    "89a65e12fdcb89f0209bb6cd1f790032f147db276000554e99600620947c665f",
			FromIndex:   0,
			FromAmount:  147690,
		})},
		TxOuts: []bo.BchTxOutTpl{bo.BchTxOutTpl{
			ToAddr:   "qrxny22vj967nygks5fgm33r6984cl7zlu9h5madq3",
			ToAmount: 137000,
		}, bo.BchTxOutTpl{
			ToAddr:   "1KhycbX9fvG81anH5WFqGcWfu9mQvFxZpV",
			ToAmount: 690,
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
