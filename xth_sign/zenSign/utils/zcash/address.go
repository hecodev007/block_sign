package zcash

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
)

//var ZecMainet zecutil.ChainParams
var chaincfgParams = &chaincfg.Params{Name: "main"}

func init() {
	zecutil.NetList["main"] = zecutil.ChainParams{
		PubHashPrefixes:    []byte{0x20, 0x89},
		ScriptHashPrefixes: []byte{0x20, 0x96},
	}
}

const (
	UNSUPPORT = 0 //暂时不支持,标识
	P2SH      = 1 //定义P2SH地址类型
	P2PKH     = 2 //定义P2PKH地址类型
)

//抽离txscript.PayToAddrScript的方法，判断地址类型
func CheckAddressType(addr btcutil.Address) int {
	switch addr := addr.(type) {
	case *zecutil.ZecAddressPubKeyHash:
		//t1开头
		if addr == nil {
			return -1
		}
		return P2PKH
	case *zecutil.ZecAddressScriptHash:
		//t3开头
		if addr == nil {
			return -1
		}
		return P2SH
	default:
		return UNSUPPORT
	}
}

func GenAccount() (address string, private string, err error) {
	if priv, err := btcec.NewPrivateKey(btcec.S256()); err != nil {
		return "", "", err
	} else if priWif, err := btcutil.NewWIF(priv, &chaincfg.MainNetParams, true); err != nil {
		return "", "", err
	} else if address, err = zecutil.Encode(priWif.PrivKey.PubKey().SerializeCompressed(), chaincfgParams); err != nil {
		return "", "", nil
	} else {
		fmt.Println(hex.EncodeToString(priWif.PrivKey.PubKey().SerializeCompressed()))
		return address, priWif.String(), nil
	}
}
