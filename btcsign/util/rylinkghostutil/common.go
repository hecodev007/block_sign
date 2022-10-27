package rylinkghostutil

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/sirupsen/logrus"
)

var CoinNet = &chaincfg.MainNetParams

func initMain() {
	//使用主网设置
	CoinNet.PubKeyHashAddrID = 0x26
	CoinNet.ScriptHashAddrID = 0x61
	CoinNet.Bech32HRPSegwit = "ghost"
	CoinNet.Net = 0xb4eff2fb
	CoinNet.PrivateKeyID = 0xA6
}

func NewCoinNet(net string) {
	if net == "test" {
		CoinNet = &chaincfg.TestNet3Params
		logrus.Info("使用测试网络")
	} else {
		CoinNet = &chaincfg.MainNetParams
		logrus.Info("使用正式网络")
	}
}

//// Address encoding magics
//PubKeyHashAddrID:        0x00, // starts with 1
//ScriptHashAddrID:        0x05, // starts with 3
//PrivateKeyID:            0x80, // starts with 5 (uncompressed) or K (compressed)
//WitnessPubKeyHashAddrID: 0x06, // starts with p2
//WitnessScriptHashAddrID: 0x0A, // starts with 7Xh
//
//// BIP32 hierarchical deterministic extended key magics
//HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
//HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
