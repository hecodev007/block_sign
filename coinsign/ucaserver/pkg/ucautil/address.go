package ucautil

import (
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

var params = &chaincfg.Params{}

func init() {
	//main
	iniMain()
}

func iniMain() {
	//主网
	params = &chaincfg.MainNetParams
	params.PubKeyHashAddrID = 0x44
	params.ScriptHashAddrID = 0x82
	params.PrivateKeyID = 0xc0
}

func initTest() {
	//测试网

}

func CreatePrivateKey() (*btcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return btcutil.NewWIF(secret, params, true)
}

func ImportWIF(wifStr string) (*btcutil.WIF, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(params) {
		return nil, errors.New("the wif string is not valid for the bitcoin network")
	}
	return wif, nil
}

func GetAddress(wif *btcutil.WIF) (*btcutil.AddressPubKey, error) {
	return btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), params)
}

//获取Staking地址
func GetStakingAddress(wif *btcutil.WIF) (*btcutil.AddressPubKey, error) {
	id := params.PubKeyHashAddrID
	defer func() {
		params.PubKeyHashAddrID = id
	}()
	params.PubKeyHashAddrID = 0x1c
	return btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), params)
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
