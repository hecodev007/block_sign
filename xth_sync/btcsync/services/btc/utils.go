package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/o3labs/neo-utils/neoutils/btckey"
	"github.com/zwjlink/ltcd/chaincfg"
	"github.com/zwjlink/ltcutil"
)

func BytesToNeoAddr(bts string) string {
	bt, _ := hex.DecodeString(bts)
	return btckey.B58checkencodeNEO(0x17, bt)
}
func bytesToInt(he string) (int64, error) {
	if len(he)%2 == 1 {
		he = "0" + he
	}
	if len(he) > 16 {
		for i := 16; i < len(he); i++ {
			if he[i] != '0' {
				return 0, fmt.Errorf("data too long")
			}
		}
		he = he[:16]
	}

	for len(he) < 16 {
		he = he + "0"
	}
	b, err := hex.DecodeString(he)
	if err != nil {
		return 0, err
	}
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint64
	err = binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return int64(tmp), err
}

func ChangeAddrLtcToBTC(address string) (btcAddr string, err error) {
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
