package mars

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)
var NetParams *chaincfg.Params

func init(){
	NetParams = new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x32
	NetParams.ScriptHashAddrID = 0x05
	NetParams.PrivateKeyID = 0x80
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
}

func GenAccount() (address string,private string,err error){
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := wif.SerializePubKey()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	if err != nil {
		return "", "", err
	}
	address = pkhash.EncodeAddress()
	return address, wif.String(), nil
}
