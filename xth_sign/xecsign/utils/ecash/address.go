package ecash

import (
	"strings"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
)

var (
	CAddrPrefix string = "bitcoincash"
	EAddrPrefix string = "ecash"
)

//ecash addr
func GenAccount() (ecashAddrWithPrefix string, private string, err error) {
	privateKey, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return "", "", err
	}

	mainNetParams := chaincfg.MainNetParams
	mainNetParams.CashAddressPrefix = "ecash"

	wif, err := bchutil.NewWIF(privateKey, &mainNetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := (*bchec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := bchutil.NewAddressPubKeyHash(bchutil.Hash160(pk), &mainNetParams)
	if err != nil {
		return "", "", err
	}
	ecashAddrWithPrefix = "ecash:" + pkhash.EncodeAddress()

	return ecashAddrWithPrefix, wif.String(), nil
}

//1 => q to bitcoincash
func ToCashAddr(addr string) (string, error) {
	// if !strings.HasPrefix(addr, "q") {
	// 	return addr, errors.New(addr + " unsuport address")
	// }

	mainNetParams := chaincfg.MainNetParams
	mainNetParams.CashAddressPrefix = "bitcoincash"

	address, err := bchutil.DecodeAddress(addr, &mainNetParams)
	if err != nil {
		return addr, err
	}
	scaddr := address.ScriptAddress()
	pkhash, err := bchutil.NewAddressPubKeyHash(scaddr, &mainNetParams)
	if err != nil {
		return "", err
	}
	CashAddr_addr := pkhash.EncodeAddress()
	return CashAddr_addr, nil
}

//to ecash
func ToECashAddr(addr string) (string, error) {

	mainNetParams := chaincfg.MainNetParams
	mainNetParams.CashAddressPrefix = "ecash"

	address, err := bchutil.DecodeAddress(addr, &mainNetParams)
	if err != nil {
		return addr, err
	}
	scaddr := address.ScriptAddress()
	pkhash, err := bchutil.NewAddressPubKeyHash(scaddr, &mainNetParams)
	if err != nil {
		return "", err
	}
	CashAddr_addr := pkhash.EncodeAddress()
	return CashAddr_addr, nil
}

//addr: any of leg, ecash, bitcash addr
func ToECashAndCashAddr(addr string) (ecashaddr, cashaddr string, err error) {
	var (
		laddr bchutil.Address
	)

	emainNetParams := chaincfg.MainNetParams
	emainNetParams.CashAddressPrefix = "ecash"

	cmainNetParams := chaincfg.MainNetParams
	cmainNetParams.CashAddressPrefix = "bitcoincash"

	eaddr, err1 := bchutil.DecodeAddress(addr, &emainNetParams)
	caddr, err2 := bchutil.DecodeAddress(addr, &cmainNetParams)
	if err1 != nil && err2 != nil {
		return "", "", err1
	} else if nil == err1 {
		laddr = eaddr
	} else if nil == err2 {
		laddr = caddr
	}

	scaddr := laddr.ScriptAddress()

	//
	epkhash, err := bchutil.NewAddressPubKeyHash(scaddr, &emainNetParams)
	if err != nil {
		return "", "", err
	}
	ECashAddr_addr := epkhash.EncodeAddress()
	//
	cpkhash, err := bchutil.NewAddressPubKeyHash(scaddr, &cmainNetParams)
	if err != nil {
		return "", "", err
	}
	CCashAddr_addr := cpkhash.EncodeAddress()
	return ECashAddr_addr, CCashAddr_addr, nil
}

//q => 1
func CashToAddr(addr string) (string, error) {
	//ecash address fromat
	mainNetParams := chaincfg.MainNetParams
	mainNetParams.CashAddressPrefix = "bitcoincash"

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

// to ecash
func ECashToAddr(addr string) (string, error) {
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

func AllToAddr(addr string) (string, error) {
	if strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") {
		return addr, nil
	}

	addr, err := CashToAddr(addr)
	if err == nil {
		return addr, err
	}

	return ECashToAddr(addr)
}
