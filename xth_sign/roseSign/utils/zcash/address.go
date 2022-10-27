package zcash

import (
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
)

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
