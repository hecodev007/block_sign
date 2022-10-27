package rylink

import (
	"fmt"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/zwjlink/ltcd/chaincfg"
	"github.com/zwjlink/ltcutil"
)

func ChangeAddrLtcToBtc2(address string) (btcAddr string, err error) {
	addr, err := ltcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	switch addr := addr.(type) {
	case *ltcutil.AddressScriptHash:
		addrbtc, err := btcutil.NewAddressScriptHashFromHash(addr.ScriptAddress(), &btcchaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrbtc.EncodeAddress()
		return btcAddr, nil
	default:
		return "", fmt.Errorf("不支持的地址类型:%s", addr)
	}
}
