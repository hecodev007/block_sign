package rylinkdashutil

import (
	"encoding/hex"
	"fmt"
	"testing"
)

//"address": "GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG",
//"privkey": "RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"
//"pubkey": "03ed43dad72976c2b495e1ab9c6ad788e3952e397d602f7bc83453b631a24b224d"

func TestCreate(t *testing.T) {
	initMain()
	wif, _ := CreatePrivateKey()
	address, _ := GetAddress(wif)
	fmt.Println("Common Address:", address.EncodeAddress())

	//wif, _ := ImportWIF("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx")
	//address, _ := GetAddress(wif)

	fmt.Println("常规地址:", address.EncodeAddress())
	_, pub := ParsePrivKey(wif.String())
	fmt.Println("地址私钥:", wif.String())
	fmt.Println("地址公钥:", hex.EncodeToString(pub.SerializeCompressed()))

	//pubKey := wif.PrivKey.PubKey().SerializeCompressed()
	//script := pubKeyHashToScript(pubKey)
	//w, err := btcutil.NewAddressScriptHash(script, CoinNet)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("隔离见证:", w.String())

	//wif, _ := ImportWIF("your compressed privateKey Wif")
	//wif, _ := ImportWIF("L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg")
	//wif, _ := ImportWIF("L3Vi4VajAW9d6soas367pztKftX4bssE99HZkKcgj1wseJ2Wv1X5")
	//
	//address, _ := GetAddress(wif)
	//
	//fmt.Println("Common Address:", address.EncodeAddress())
	//
	//pubKey := wif.PrivKey.PubKey().SerializeCompressed()
	//
	//script := pubKeyHashToScript(pubKey)
	//w, err := btcutil.NewAddressScriptHash(script, CoinNet)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("Segregated Witness Address:", w.String())
	////fmt.Println(" Witness Address:", s.String())
	//fmt.Println("PrivateKeyWifCompressed:", wif.String())
	//
	//addddr, _ := btcutil.DecodeAddress("362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU", CoinNet)
	//fmt.Println("addddr:", addddr.String())
	//addrBytes, _ := txscript.PayToAddrScript(addddr)
	//scriptPubKey := hex.EncodeToString(addrBytes)
	//fmt.Println("addddr scriptPubKey", scriptPubKey)
	//
	//addddr2, _ := btcutil.DecodeAddress("18q3NwcLQ64xvE1phEn9os2Z3C1L9hUme1", CoinNet)
	//fmt.Println("addddr2:", addddr2.String())
	//addrBytes2, _ := txscript.PayToAddrScript(addddr2)
	//scriptPubKey2 := hex.EncodeToString(addrBytes2)
	//fmt.Println("addddr2 scriptPubKey", scriptPubKey2)
	//btcutil.NewAddressPubKey()
	//testpk,_ := btcutil.NewAddressPubKey(addrBytes2,params)
	//fmt.Println(testpk.String())

}

func TestCreateAddr(t *testing.T) {
	t.Log(CreateAddr())
}
