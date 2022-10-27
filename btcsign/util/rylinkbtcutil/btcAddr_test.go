package rylinkbtcutil

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"testing"
)

func TestCreate(t *testing.T) {
	//wif, _ := CreatePrivateKey()
	//wif, _ := ImportWIF("your compressed privateKey Wif")
	//wif, _ := ImportWIF("L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg")
	//wif, _ := ImportWIF("L3Vi4VajAW9d6soas367pztKftX4bssE99HZkKcgj1wseJ2Wv1X5")
	wif, _ := ImportWIF("L273HNzNDcU4WHkSqVEv8sRie3HdAAxkWzwgrNLNaCPcgKMW55Lu")
	//fmt.Println(hex.EncodeToString(wif.PrivKey.D.Bytes()))
	//return
	address, _ := GetAddress(wif)

	fmt.Println("Common Address:", address.EncodeAddress())

	pubKey := wif.PrivKey.PubKey().SerializeCompressed()

	script := pubKeyHashToScript(pubKey)
	w, err := btcutil.NewAddressScriptHash(script, CoinNet)
	if err != nil {
		panic(err)
	}

	fmt.Println("Segregated Witness Address:", w.String())
	//fmt.Println(" Witness Address:", s.String())
	fmt.Println("PrivateKeyWifCompressed:", wif.String())

	addddr, _ := btcutil.DecodeAddress("362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU", CoinNet)
	fmt.Println("addddr:", addddr.String())
	addrBytes, _ := txscript.PayToAddrScript(addddr)
	scriptPubKey := hex.EncodeToString(addrBytes)
	fmt.Println("addddr scriptPubKey", scriptPubKey)

	addddr2, _ := btcutil.DecodeAddress("18q3NwcLQ64xvE1phEn9os2Z3C1L9hUme1", CoinNet)
	fmt.Println("addddr2:", addddr2.String())
	addrBytes2, _ := txscript.PayToAddrScript(addddr2)
	scriptPubKey2 := hex.EncodeToString(addrBytes2)
	fmt.Println("addddr2 scriptPubKey", scriptPubKey2)
	//btcutil.NewAddressPubKey()
	//testpk,_ := btcutil.NewAddressPubKey(addrBytes2,params)
	//fmt.Println(testpk.String())

	//bc开头地址
	pk, err := txscript.ParsePkScript(script)
	if err != nil {
		panic(err)
	}
	apk, err := pk.Address(CoinNet)
	fmt.Println(apk.EncodeAddress())

	// 得到隔离见证地址的回执脚本
	//const redeemScript = bitcoin.script.witnessPubKeyHash.output.encode(pubKeyHash);
}

func TestCreateAddr(t *testing.T) {
	t.Log(CreateAddr())
}
