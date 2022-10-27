package model

import (
	"testing"
	"time"
)

func Test_scan(t *testing.T) {
	txs, err := new(Scan).BlockByHeight(730525)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(len(txs))
}

func Test_trx_scan(t *testing.T) {
	amount, tm, err := new(TrxScan).BalanceOf("TBdiKcS1rdzasufqnNrupZJMPmAni9oyYn", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(amount)
	et := time.Unix(tm, 0)
	t.Log(et.String())
}

func Test_eth_scan(t *testing.T) {
	t.Log(new(EthNode).BlockCount())
	amount, tm, err := new(EthNode).BalanceOf("0x00ac94261a72a79bef4af7b3d93bc01a735e7212", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(amount)
	et := time.Unix(tm, 0)
	t.Log(et.String())
}
