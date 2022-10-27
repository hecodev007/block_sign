package dash

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	dashchaincfg "github.com/dashpay/godash/chaincfg"
	"github.com/dashpay/godashutil"
)

var NetParams *chaincfg.Params
var DashUtilNetParams *dashchaincfg.Params

func init() {
	//from:https://github.com/dashpay/dash/src/chainparams.cpp
	NetParams = new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x4c
	NetParams.ScriptHashAddrID = 0x10
	NetParams.PrivateKeyID = 0xcc
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	DashUtilNetParams = new(dashchaincfg.Params)
	DashUtilNetParams.PubKeyHashAddrID = 0x4c
	DashUtilNetParams.ScriptHashAddrID = 0x10
	DashUtilNetParams.PrivateKeyID = 0xcc
	DashUtilNetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	DashUtilNetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
}
func GenAccount() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := (*btcec.PublicKey)(&pri.PublicKey).SerializeCompressed()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	if err != nil {
		return "", "", err
	}
	address = pkhash.EncodeAddress()
	return address, wif.String(), nil
}
func DecodeAddress(addr string) (godashutil.Address, error) {
	return godashutil.DecodeAddress(addr, DashUtilNetParams)
}
