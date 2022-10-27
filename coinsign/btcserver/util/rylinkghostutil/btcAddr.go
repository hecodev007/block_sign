package rylinkghostutil

import (
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

func CreatePrivateKey() (*btcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return btcutil.NewWIF(secret, CoinNet, true)
}

func ImportWIF(wifStr string) (*btcutil.WIF, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(CoinNet) {
		return nil, errors.New("the wif string is not valid for the bitcoin network")
	}
	return wif, nil
}

func GetAddress(wif *btcutil.WIF) (*btcutil.AddressPubKey, error) {
	return btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), CoinNet)
}

func pubKeyHashToScript(pubKey []byte) []byte {
	pubKeyHash := btcutil.Hash160(pubKey)
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).AddData(pubKeyHash).Script()
	if err != nil {
		panic(err)
	}
	return script
}

func pubKeyHashToScript2(pubKey []byte) []byte {
	pubKeyHash := btcutil.Hash160(pubKey)
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).AddData(pubKeyHash).Script()
	if err != nil {
		panic(err)
	}
	return script
}

func CreateAddr() (addr, segwAddr, bcSegwAddr, privkey string, err error) {
	wif, _ := CreatePrivateKey()
	address, err := GetAddress(wif)
	if err != nil {
		return "", "", "", "", err
	}
	addr = address.EncodeAddress()
	privkey = wif.String()

	pubKey := wif.PrivKey.PubKey().SerializeCompressed()
	script := pubKeyHashToScript(pubKey)
	w, err := btcutil.NewAddressScriptHash(script, CoinNet)
	if err != nil {
		return "", "", "", "", err
	}
	segwAddr = w.String()

	//bc开头地址
	pk, err := txscript.ParsePkScript(script)
	if err != nil {
		return "", "", "", "", err
	}
	apk, err := pk.Address(CoinNet)
	//fmt.Println(apk.EncodeAddress())
	bcSegwAddr = apk.EncodeAddress()
	return
}
