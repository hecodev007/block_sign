package bcha

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchutil"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	ad "github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/txscript"
)

func Test_addr(t *testing.T) {
	//laddr, caddr, pri, err := GenAccount()
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//t.Log(laddr, caddr, pri)
	caddr := "1AnJeoPNk4yzRv5GqoqusbpEtbmLzVJteN"
	addr, err := ToCashAddr(caddr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)

	caddr, err = ToAddr(addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(caddr)
}
func Test_script(t *testing.T) {
	script, err := hex.DecodeString("a914260617ebf668c9102f71ce24aba97fcaaf9c666a87")
	if err != nil {
		t.Fatal(err.Error())
	}
	_, addrs, num, err := txscript.ExtractPkScriptAddrs(script, &ad.MainNetParams)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(num, addrs[0].EncodeAddress())
}

func Test_uencaddr2(t *testing.T) {
	privateKey, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
	}

	wif, err := bchutil.NewWIF(privateKey, &ad.MainNetParams, false)
	if err != nil {
	}
	pk := (*bchec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := bchutil.NewAddressPubKeyHash(bchutil.Hash160(pk), &ad.MainNetParams)
	if err != nil {
	}
	CashAddr_addr := pkhash.EncodeAddress()
	LegacyAddressPubKeyHash, err := bchutil.NewLegacyAddressPubKeyHash(bchutil.Hash160(pk), &ad.MainNetParams)
	if err != nil {
	}
	Legacy_addr := LegacyAddressPubKeyHash.EncodeAddress()
	fmt.Println(Legacy_addr)
	fmt.Println(CashAddr_addr)
	fmt.Println(hex.EncodeToString(wif.PrivKey.D.Bytes()))

}
func Test_uencaddr(t *testing.T) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		fmt.Println(err)
	}
	privKeyWif, err := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, true)
	if err != nil {
		fmt.Println(err)
	}

	pubKeySerial := privKey.PubKey().SerializeCompressed()
	pubKeyAddress, err := btcutil.NewAddressPubKey(pubKeySerial, &chaincfg.MainNetParams)

	pubKeySerial1 := privKey.PubKey().SerializeUncompressed()
	pubKeyAddress1, err := btcutil.NewAddressPubKey(pubKeySerial1, &chaincfg.MainNetParams)

	pubKeySerial2 := privKey.PubKey().SerializeHybrid()
	pubKeyAddress2, err := btcutil.NewAddressPubKey(pubKeySerial2, &chaincfg.MainNetParams)

	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pubKeyAddress.EncodeAddress())
	fmt.Println(pubKeyAddress1.EncodeAddress())
	fmt.Println(pubKeyAddress2.EncodeAddress())
	fmt.Println(hex.EncodeToString(privKeyWif.PrivKey.D.Bytes()))

	//var pri ecdsa.PrivateKey
	//pri.D, _ = new(big.Int).SetString("E83385AF76B2B1997326B567461FB73DD9C27EAB9E1E86D26779F4650C5F2B75",16)
	//pri.PublicKey.Curve = elliptic.P256()
	//pri.PublicKey.X, pri.PublicKey.Y = pri.PublicKey.Curve.ScalarBaseMult(pri.D.Bytes())

	//privateKey := "deb7456cbc280d7d1a2a90d90dfd0a0c6ca95d7214ad32cf08f14f107ab53e3c"
	//log.Println(strings.ToUpper(Public(privateKey)))

}

func Public(privateKey string) (publicKey string) {
	var e ecdsa.PrivateKey
	e.D, _ = new(big.Int).SetString(privateKey, 16)
	e.PublicKey.Curve = secp256k1.S256()
	e.PublicKey.X, e.PublicKey.Y = e.PublicKey.Curve.ScalarBaseMult(e.D.Bytes())
	return string(Base58Encode(elliptic.Marshal(secp256k1.S256(), e.X, e.Y)))
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0
	for _, b := range input {
		if b != b58Alphabet[0] {
			break
		}
		zeroBytes++
	}
	payload := input[zeroBytes:]
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(int64(len(b58Alphabet))))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}

func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	ReverseBytes(result)

	for _, b := range input {
		if b == 0x00 {
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break
		}
	}
	return result

}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
