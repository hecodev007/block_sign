package rylinkbtcutil

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/sirupsen/logrus"
)

var CoinNet = &chaincfg.MainNetParams

func NewCoinNet(net string) {
	if net == "test" {
		CoinNet = &chaincfg.TestNet3Params
		logrus.Info("使用测试网络")
	} else {
		CoinNet = &chaincfg.MainNetParams
		logrus.Info("使用正式网络")
	}
}
