package main

import (
	"btcont/common/conf"
	"btcont/common/model"
	"testing"

	"github.com/shopspring/decimal"
)

func Test_node(t *testing.T) {
	node := model.NewEthNode(conf.Cfg.Nodes[COINNAME])
	a, tm, err := node.BalanceOf("0x05f0fdd0e49a5225011fff92ad85cc68e1d1f08e", "0x7c63f96feafacd84e75a594c00fac3693386fbf0")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(tm / 360)
	t.Log(a.Shift(-9).String())
}

func Test_dec(t *testing.T) {
	d, err := decimal.NewFromString("")
	t.Log(err.Error())
	t.Log(d.Cmp(decimal.Zero))
}
