package btc

import (
	"errors"
	"fmt"

	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
)

const (
	TX_NONSTANDARD           = "nonstandard"
	TX_PUBKEY                = "pubkey"
	TX_PUBKEYHASH            = "pubkeyhash"
	TX_SCRIPTHASH            = "scripthash"
	TX_MULTISIG              = "multisig"
	TX_NULL_DATA             = "nulldata"
	TX_WITNESS_V0_KEYHASH    = "witness_v0_keyhash"
	TX_WITNESS_V0_SCRIPTHASH = "witness_v0_scripthash"
	TX_WITNESS_UNKNOWN       = "witness_unknown"
)

type scriptPubkey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

func (sp *scriptPubkey) GetAddress() ([]string, error) {
	if sp.Addresses == nil || len(sp.Addresses) == 0 {
		return nil, errors.New("empty out.Address")
	}
	switch sp.Type {
	case TX_PUBKEY, TX_PUBKEYHASH, TX_SCRIPTHASH, TX_WITNESS_V0_KEYHASH, TX_WITNESS_V0_SCRIPTHASH:

		return sp.Addresses, nil
		// addrs := make([]string, 0)
		// for _, addr := range sp.Addresses {
		// 	laddr, _ := ECashToLegacyAddr(addr)
		// 	addrs = append(addrs, laddr)
		// }
		// return addrs, nil
	default:
		return nil, fmt.Errorf("don't support tx %s", sp.Type)
	}
}

func ECashToLegacyAddr(addr string) (string, error) {
	//ecash address fromat
	mainNetParams := chaincfg.MainNetParams
	mainNetParams.CashAddressPrefix = "ecash"

	address, err := bchutil.DecodeAddress(addr, &mainNetParams)
	if err != nil {
		return addr, err
	}
	scaddr := address.ScriptAddress()
	pkhash, err := bchutil.NewLegacyAddressPubKeyHash(scaddr, &mainNetParams)
	if err != nil {
		return "", err
	}
	CashAddr_addr := pkhash.EncodeAddress()
	return CashAddr_addr, nil
}
