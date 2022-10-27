package btc

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/btcec"

	"testing"
)
var SatParams *chaincfg.Params

func init(){
	SatParams = new(chaincfg.Params)
	SatParams.PubKeyHashAddrID = 0x3f
	SatParams.ScriptHashAddrID = 0x41
	SatParams.PrivateKeyID = 0x1e
	SatParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	SatParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	SatParams.Bech32HRPSegwit = "sat"
}

func Test_acc(t *testing.T){
	pri := "5ZBjUDZnXUXLAuAmwxJ8qstmAqiEwPQrNjfJDy4bWZ3uZni3bmDn"
	//pri = "KwSVmzdUqcGd5wXW43Ya8geyTGAM9NxyLGnmdjK4m4a5Nm8B2F8i"
	wif,err :=btcutil.DecodeWIF(pri)
	if err != nil {
		panic(err.Error())
	}
	pribytes:=wif.PrivKey.Serialize()
	privatekey,pubkey := btcec.PrivKeyFromBytes(btcec.S256(),pribytes)
	_=privatekey
	t.Log(hex.EncodeToString(pribytes))
	t.Log(pubkey.SerializeCompressed())
	t.Log(wif.SerializePubKey())

	t.Log(wif.CompressPubKey)
	pkhash,err :=btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.SerializePubKey()),SatParams)
	if err != nil{
		panic(err.Error())
	}
	address := pkhash.EncodeAddress()
	fmt.Println(address)
	//SNv23NJ3HgYhWo38YTmyVV1PMSE1eXdp8G
	//SNv23NJ3HgYhWo38YTmyVV1PMSE1eXdp8G
}
func Test_gen(t *testing.T){
	addr,pri,err := GenAccount2()
	t.Log(addr,pri,err)
	pribytes ,_ := hex.DecodeString("ca98010aab7c0eaedb038c8fc21dc1b46917d5d37938b9e136c86d9f10cdfc32")
	prikey,_ :=btcec.PrivKeyFromBytes(btcec.S256(),pribytes)
	wif, err := btcutil.NewWIF(prikey, SatParams, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	pk := wif.SerializePubKey()

	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), SatParams)
	if err != nil {
		t.Fatal(err.Error())
	}
	address := pkhash.EncodeAddress()
	t.Log( address, wif.String())
}
func GenAccount2() (address string,private string,err error){
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, SatParams, true)
	if err != nil {
		return "", "", err
	}
	pk := wif.SerializePubKey()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), SatParams)
	if err != nil {
		return "", "", err
	}
	address = pkhash.EncodeAddress()
	return address, wif.String(), nil
}

//sat1qu0lulq5mxrsqtk6zl3jeq78enmyafv2y7ntju5
//12Dni1tZ6E6DPtPAmfQ9ey5381ZexVGRvA
//1qu0lulq5mxrsqtk6zl3jeq78enmyafv2y7ntju5
//1qeee050gxncpy0k2k8c6lwgtrutrh537gkp3edw
//sat1qeee050gxncpy0k2k8c6lwgtrutrh537gkp3edw
//sat1qu0lulq5mxrsqtk6zl3jeq78enmyafv2y7ntju5
//ScZfhnfe1bWNHUHvMNTRgPQfNyHGy4ndPR