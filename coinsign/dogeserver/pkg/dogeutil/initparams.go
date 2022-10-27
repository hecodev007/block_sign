package dogeutil

import (
	"github.com/btcsuite/btcd/chaincfg"
)

var coinNetPrams = &chaincfg.Params{}
var dogeDecimal int32 = 8

func init() {
	//main
	iniMain()
}

func iniMain() {
	//主网
	coinNetPrams = &chaincfg.MainNetParams
	coinNetPrams.PubKeyHashAddrID = 0x1e
	coinNetPrams.ScriptHashAddrID = 0x16
	coinNetPrams.PrivateKeyID = 0x9e
	//coinNetPrams.WitnessPubKeyHashAddrID = 0x00
	//coinNetPrams.WitnessScriptHashAddrID = 0x0A
	//coinNetPrams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	//coinNetPrams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}

}

func initTest() {
	//测试网

}
