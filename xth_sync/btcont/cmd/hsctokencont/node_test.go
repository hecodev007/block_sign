package main

import (
	"btcont/common/conf"
	"btcont/common/model"
	"testing"
)

func Test_node(t *testing.T) {
	node := model.NewEthNode(conf.Cfg.Nodes["eth"])
	a, _, err := node.BalanceOf("0x0058174f72050846eeec6feec1daa233e2e71c2e", "0x7db5af2B9624e1b3B4Bb69D6DeBd9aD1016A58Ac")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(a.Shift(-9).String())
}
