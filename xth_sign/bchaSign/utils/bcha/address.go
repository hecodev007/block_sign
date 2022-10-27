package bcha

import (
	"errors"
	"strings"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
)

func GenAccount() (Legacy_addr string, CashAddr_addr string, private string, err error) {
	privateKey, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return "", "", "", err
	}

	wif, err := bchutil.NewWIF(privateKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", "", err
	}
	pk := (*bchec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := bchutil.NewAddressPubKeyHash(bchutil.Hash160(pk), &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", err
	}
	CashAddr_addr = pkhash.EncodeAddress()
	LegacyAddressPubKeyHash, err := bchutil.NewLegacyAddressPubKeyHash(bchutil.Hash160(pk), &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", err
	}
	Legacy_addr = LegacyAddressPubKeyHash.EncodeAddress()
	return Legacy_addr, CashAddr_addr, wif.String(), nil
}

//q => 1
func ToCashAddr(addr string) (string, error) {
	if strings.HasPrefix(addr, "q") {
		return addr, nil
	}
	if !strings.HasPrefix(addr, "1") {
		return addr, errors.New(addr + " unsuport address")
	}
	address, err := bchutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		return addr, err
	}
	scaddr := address.ScriptAddress()
	pkhash, err := bchutil.NewAddressPubKeyHash(scaddr, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	CashAddr_addr := pkhash.EncodeAddress()
	return CashAddr_addr, nil
}

//1 => q
func ToAddr(addr string) (string, error) {
	if strings.HasPrefix(addr, "1") {
		return addr, nil
	}
	if !strings.HasPrefix(addr, "q") {
		return addr, errors.New(addr + " unsuport address")
	}
	address, err := bchutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		return addr, err
	}
	scaddr := address.ScriptAddress()
	pkhash, err := bchutil.NewLegacyAddressPubKeyHash(scaddr, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	CashAddr_addr := pkhash.EncodeAddress()
	return CashAddr_addr, nil
}
