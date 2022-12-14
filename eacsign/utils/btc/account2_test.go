package btc

import (

	//"crypto/elliptic"
	//"crypto/rand"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/status-im/keycard-go/hexutils"
	"testing"
)

//func GeneratePublicKeyHash2(publicKey []byte) []byte {
//	sha256PubKey := sha1.Sum(publicKey)
//	r := ripemd160.New()
//	r.Write(sha256PubKey[:])
//	ripPubKey := r.Sum(nil)
//	return ripPubKey
//}

func Test_acc2(t *testing.T) {
	NetParams := new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x3f
	NetParams.ScriptHashAddrID = 0x41
	NetParams.PrivateKeyID = 0x1e
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	NetParams.Bech32HRPSegwit = "sat"

	pri := "L4SdL6gRUfDfDGg2wDptnPxXgacWcRuWq5xmUwnxWu2fjpagyiwg"
	//pri = "KwSVmzdUqcGd5wXW43Ya8geyTGAM9NxyLGnmdjK4m4a5Nm8B2F8i"
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		panic(err.Error())
	}
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), NetParams)
	if err != nil {
		panic(err.Error())
	}
	address := pkhash.EncodeAddress()
	fmt.Println(address)
	return
}

func Test_sat22(t *testing.T) {
	NetParams := new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x3f
	NetParams.ScriptHashAddrID = 0x41
	NetParams.PrivateKeyID = 0x1e
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	NetParams.Bech32HRPSegwit = "sat"
	wif, err2 := btcutil.DecodeWIF("5SstwMY3o2iaj1HZ2T63Fma7xyq24g5x44NqucdyWcpEtWVe5m6y")

	if err2 != nil {
		fmt.Println(err2.Error())
	}

	//pri, err := btcec.NewPrivateKey(btcec.S256())
	//wif, err := btcutil.NewWIF(pri, NetParams, true)
	pk := (*btcec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	address2 := pkhash.EncodeAddress()
	conv, err := bech32.ConvertBits(btcutil.Hash160(pk), 8, 5, true)
	if err != nil {
		fmt.Println("Error:", err)
	}
	versionPlusData := make([]byte, 1+len(conv))
	versionPlusData[0] = 0
	for i, d := range conv {
		versionPlusData[i+1] = d
	}
	address, err := bech32.Encode("sat", versionPlusData)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("??????: ", address)
	fmt.Println("??????2: ", address2)
	fmt.Println("??????: ", wif.String())
}

//TKkuS81difdfDBKq6MBnTSTQ4B5r75PSWp,6479d5b583dc0e695c809e18124ca893c6efa7d59ce1b43feb61acd35806000c

func Test_sat222(t *testing.T) {
	NetParams := new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x3f
	NetParams.ScriptHashAddrID = 0x41
	NetParams.PrivateKeyID = 0x1e
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	NetParams.Bech32HRPSegwit = "sat"

	bytes := hexutils.HexToBytes("6479d5b583dc0e695c809e18124ca893c6efa7d59ce1b43feb61acd35806000c")
	pri, _ := btcec.PrivKeyFromBytes(btcec.S256(), bytes)
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	pk := (*btcec.PublicKey)(&pri.PublicKey).SerializeCompressed()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	address2 := pkhash.EncodeAddress()

	conv, err := bech32.ConvertBits(btcutil.Hash160(pk), 8, 5, true)
	if err != nil {
		fmt.Println("Error:", err)
	}
	versionPlusData := make([]byte, 1+len(conv))
	versionPlusData[0] = 0
	for i, d := range conv {
		versionPlusData[i+1] = d
	}
	address, err := bech32.Encode("sat", versionPlusData)
	if err != nil {
		fmt.Println("Error:", err)
	}

	hash, err := btcutil.NewAddressScriptHash(pkhash.ScriptAddress(), NetParams)
	fmt.Println("??????: ", address)
	fmt.Println("??????2: ", address2)
	fmt.Println("??????3: ", hash.EncodeAddress())
	fmt.Println("??????: ", wif.String())
}

func Test_decodedAddress2(t *testing.T) {
	NetParams := new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x3f
	NetParams.ScriptHashAddrID = 0x41
	NetParams.PrivateKeyID = 0x1e
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	NetParams.Bech32HRPSegwit = "sat"
	decode, bytes, err := bech32.Decode("sat1q37pnjwm7dsdat7zy5z5gq6z7txqmjplvv7v538")
	fmt.Println(decode)
	fmt.Println(err)
	bits, err := bech32.ConvertBits(bytes[1:], 5, 8, true)
	pkhash, err := btcutil.NewAddressPubKeyHash(bits, NetParams)
	address2 := pkhash.EncodeAddress()
	fmt.Println(address2)

}
