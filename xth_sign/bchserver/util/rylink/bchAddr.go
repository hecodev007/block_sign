package rylink

import (
	"fmt"
	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchutil"
)

func createPrivateKey() (*bchutil.WIF, error) {
	secret, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return nil, err
	}
	return bchutil.NewWIF(secret, &chaincfg.MainNetParams, true)
}

func importWIF(wifStr string) (*bchutil.WIF, error) {
	wif, err := bchutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	//if !wif.IsForNet(params) {
	//	return nil, errors.New("the wif string is not valid for the bitcoin network")
	//}
	return wif, nil
}

func getBtcP2PKHAddress(wif *bchutil.WIF) (*bchutil.LegacyAddressPubKeyHash, error) {
	pubKeyHash := bchutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	return bchutil.NewLegacyAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
}
func getBtcP2SHAddress(wif *bchutil.WIF) (*bchutil.LegacyAddressScriptHash, error) {
	pubKeyHash := pubKeyHashToScript(wif.PrivKey.PubKey().SerializeCompressed())
	return bchutil.NewLegacyAddressScriptHash(pubKeyHash, &chaincfg.MainNetParams)
}
func getBchP2PKHAddress(wif *bchutil.WIF) (*bchutil.AddressPubKeyHash, error) {
	pubKeyHash := bchutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	return bchutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
}
func getBchP2SHAddress(wif *bchutil.WIF) (*bchutil.AddressScriptHash, error) {
	pubKeyHash := pubKeyHashToScript(wif.PrivKey.PubKey().SerializeCompressed())
	return bchutil.NewAddressScriptHash(pubKeyHash, &chaincfg.MainNetParams)
}

func pubKeyHashToScript(pubKey []byte) []byte {
	pubKeyHash := bchutil.Hash160(pubKey)
	script, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(pubKeyHash).Script()
	if err != nil {
		panic(err)
	}
	return script
}

//生成地址,(p2pkh),由于冷签名尚未支持p2sh,暂时不生成p2sh地址
func CreateAddress() (bchAddr, btcAddr, privKey string, e error) {
	wif, _ := createPrivateKey()
	btcAddress, err := getBtcP2PKHAddress(wif)
	if err != nil {
		return "", "", "", err
	}
	btcAddr = btcAddress.String()
	privKey = wif.String()
	bchAddress, err := getBchP2PKHAddress(wif)
	if err != nil {
		return "", "", "", err
	}
	bchAddr = bchAddress.String()
	return
}

//生成地址,(p2pkh),由于冷签名尚未支持p2sh,暂时不生成p2sh地址
func CreateAddressP2sh() (bchAddr, btcAddr, segwBchAddr, segwBtcAddr, privKey string, e error) {
	wif, _ := createPrivateKey()
	btcAddress, err := getBtcP2PKHAddress(wif)
	if err != nil {
		return "", "", "", "", "", err
	}
	btcAddr = btcAddress.String()
	privKey = wif.String()

	segwBtcAddress, err := getBtcP2SHAddress(wif)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwBtcAddr = segwBtcAddress.String()

	bchAddress, err := getBchP2PKHAddress(wif)
	if err != nil {
		return "", "", "", "", "", err
	}
	bchAddr = bchAddress.String()

	bchSegwAddr, err := getBchP2SHAddress(wif)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwBchAddr = bchSegwAddr.String()
	return
}

//BCH转换成地址BTC
func ChangeAddressToBtc(address string) (btcAddr string, err error) {
	err = nil
	addr, err := bchutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	switch addr.(type) {
	case *bchutil.AddressPubKeyHash:
		addrbtc, err := bchutil.NewLegacyAddressPubKeyHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrbtc.EncodeAddress()
		return btcAddr, nil
	case *bchutil.AddressScriptHash:
		addrbtc, err := bchutil.NewLegacyAddressScriptHashFromHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrbtc.EncodeAddress()
		return btcAddr, nil
	//case *bchutil.AddressPubKey:
	//	return "", fmt.Errorf("不支持的地址类型:%s", addr)
	default:
		return "", fmt.Errorf("不支持的地址类型:%s", addr)
	}
}

//BCH转换成地址BTC
func ChangeAddressToBch(address string) (btcAddr string, err error) {
	err = nil
	addr, err := bchutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	switch addr.(type) {
	case *bchutil.AddressPubKeyHash:
		return address, nil
	case *bchutil.AddressScriptHash:
		return address, nil
	case *bchutil.LegacyAddressPubKeyHash:
		addrbtc, err := bchutil.NewAddressPubKeyHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrbtc.EncodeAddress()
		return btcAddr, nil
	case *bchutil.LegacyAddressScriptHash:
		addrbtc, err := bchutil.NewAddressScriptHashFromHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrbtc.EncodeAddress()
		return btcAddr, nil
	default:
		return "", fmt.Errorf("不支持的地址类型:%s", addr)
	}
}

//普通地址和隔离见证地址生成
func TestAddr() {
	btcaddr, bchaddr, privkey, _ := CreateAddress()
	fmt.Println(btcaddr, bchaddr, privkey)
	fmt.Println(ChangeAddressToBtc(bchaddr))
	fmt.Println(ChangeAddressToBch(btcaddr))
}
