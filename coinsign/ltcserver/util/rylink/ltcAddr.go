package rylink

import (
	"fmt"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/zwjlink/ltcd/btcec"
	"github.com/zwjlink/ltcd/chaincfg"
	"github.com/zwjlink/ltcd/txscript"
	"github.com/zwjlink/ltcutil"
)

func createPrivateKey() (*ltcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return ltcutil.NewWIF(secret, &chaincfg.MainNetParams, true)
}

func ImportWIF(wifStr string) (*ltcutil.WIF, error) {
	wif, err := ltcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	//if !wif.IsForNet(params) {
	//	return nil, errors.New("the wif string is not valid for the bitcoin network")
	//}
	return wif, nil
}

func getAddress(wif *ltcutil.WIF) (*ltcutil.AddressPubKey, error) {
	return ltcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.MainNetParams)
}

func pubKeyHashToScript(pubKey []byte) []byte {
	pubKeyHash := ltcutil.Hash160(pubKey)
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).AddData(pubKeyHash).Script()
	if err != nil {
		panic(err)
	}
	return script
}

//生成地址,返回普通地址，隔离见证地址和私钥
func CreateAddress() (addr, segWitAddress, privKey string, e error) {
	wif, err := createPrivateKey()
	if err != nil {
		return "", "", "", err
	}
	address, err := getAddress(wif)
	if err != nil {
		return "", "", "", err
	}
	pubKey := wif.PrivKey.PubKey().SerializeCompressed()
	script := pubKeyHashToScript(pubKey)
	w, err := ltcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", err
	}
	addr = address.EncodeAddress()
	segWitAddress = w.String()
	privKey = wif.String()
	return
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

func ChangeAddrBtcToLtc(address string) (btcAddr string, err error) {
	addr, err := btcutil.DecodeAddress(address, &btcchaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	switch addr := addr.(type) {
	case *btcutil.AddressScriptHash:
		addrLtc, err := ltcutil.NewAddressScriptHashFromHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		btcAddr = addrLtc.EncodeAddress()
		return btcAddr, nil
	default:
		return "", fmt.Errorf("不支持的地址类型:%s", addr)
	}
}

//普通地址和隔离见证地址生成
func testmain2() {
	wif, _ := createPrivateKey()
	address, _ := getAddress(wif)
	fmt.Println("Common Address:", address.EncodeAddress())

	pubKey := wif.PrivKey.PubKey().SerializeCompressed()
	script := pubKeyHashToScript(pubKey)
	w, err := ltcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("Segregated Witness Address:", w.String())
	//fmt.Println(" Witness Address:", s.String())
	fmt.Println("PrivateKeyWifCompressed:", wif.String())

}
