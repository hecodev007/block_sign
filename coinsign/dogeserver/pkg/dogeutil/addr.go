package dogeutil

import (
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

func CreatePrivateKey() (*btcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return btcutil.NewWIF(secret, coinNetPrams, true)
}

func ImportWIF(wifStr string) (*btcutil.WIF, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(coinNetPrams) {
		return nil, errors.New("the wif string is not valid for the bitcoin network")
	}
	return wif, nil
}

func GetAddress(wif *btcutil.WIF) (*btcutil.AddressPubKey, error) {
	return btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), coinNetPrams)
}

func CreateAddr() (addr, privkey string, err error) {
	wif, _ := CreatePrivateKey()
	address, err := GetAddress(wif)
	if err != nil {
		return "", "", err
	}
	addr = address.EncodeAddress()
	privkey = wif.String()
	return
}
