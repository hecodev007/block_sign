package btc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/ripemd160"
	"math/big"

	//"crypto/elliptic"
	//"crypto/rand"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/bech32"
	"testing"
)

func ByteToString(b []byte) (s string) {
	s = ""
	for i := 0; i < len(b); i++ {
		s += fmt.Sprintf("%02X", b[i])
	}
	return s
}

const VERSION = byte(0x00)

func Test_accuenc(t *testing.T) {
	reader := rand.Reader
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), reader)
	//privateKey, err := crypto.GenerateKey()
	if err != nil {
		return
	}

	var sPriStr []byte
	for i := len(privateKey.D.Bytes()) - 1; i >= 0; i-- {
		sPriStr = append(sPriStr, privateKey.D.Bytes()[i])
	}

	var sPubStr []byte
	c1 := privateKey.PublicKey.X.Bytes()
	c2 := privateKey.PublicKey.Y.Bytes()

	for i := len(c1) - 1; i >= 0; i-- {
		sPubStr = append(sPubStr, c1[i])
	}

	for i := len(c2) - 1; i >= 0; i-- {
		sPubStr = append(sPubStr, c2[i])
	}
	priSize := len(sPriStr)
	pubSize := len(sPubStr)

	fmt.Println(sPriStr)
	fmt.Println(sPubStr)
	fmt.Println(priSize)
	fmt.Println(pubSize)

	//rWoDdxeS5w52zJIcH5VSBMCFLgpv1PgK/vJ1m5Vkgs8=
	//MCZkawU14elEgmQJmaffM1GAwYipE5U2wUbo4ke4y1t08Is1u3uTbZBk9hr85cbFS4gkqI6tccvdFRFh+mhQPw==

	//"address": "1LLR9QALc7Bduh3uoUQcux37RNbc85UL2j",
	//"private_key": "5vl5oeKkX1L0GOhQ2UsxaS3qKt+y95Vy6+cm7Zibx3c=",
	//"public_key": "ICCXdvx7l3mzbWDR0RmvttDe21/kq7c+212sz2Fbi/JF6aYtVYlw340GXlGlct1yK4sx3KJ50crc7t+vvl/NtC15"

	fmt.Println(base64.StdEncoding.EncodeToString(sPriStr))
	fmt.Println(hex.EncodeToString(sPriStr))
	fmt.Println(hex.EncodeToString(privateKey.D.Bytes()))
	fmt.Println(base64.StdEncoding.EncodeToString(sPubStr))

	ripPubKey := GeneratePublicKeyHash1(sPubStr)
	//2.最前面添加一个字节的版本信息获得 versionPublickeyHash
	versionPublickeyHash := append([]byte{VERSION}, ripPubKey[:]...)
	//3.sha256(sha256(versionPublickeyHash))  取最后四个字节的值
	tailHash := CheckSumHash1(versionPublickeyHash)
	//4.拼接最终hash versionPublickeyHash + checksumHash
	finalHash := append(versionPublickeyHash, tailHash...)
	//进行base58加密
	address := Base58Encode(finalHash)
	fmt.Println(string(address))

	//ripPubKey2 := GeneratePublicKeyHash2(sPubStr)
	////2.最前面添加一个字节的版本信息获得 versionPublickeyHash
	//versionPublickeyHash2 := append([]byte{VERSION}, ripPubKey2[:]...)
	////3.sha256(sha256(versionPublickeyHash))  取最后四个字节的值
	//tailHash2 := CheckSumHash2(versionPublickeyHash2)
	////4.拼接最终hash versionPublickeyHash + checksumHash
	//finalHash2 := append(versionPublickeyHash2, tailHash2...)
	////进行base58加密
	//address2 := Base58Encode(finalHash2)
	//fmt.Println(string(address2))

	//Private, err := x509.MarshalECPrivateKey(privateKey)
	//public := privateKey.PublicKey
	////x509 serialization
	//publicKey, err := x509.MarshalPKIXPublicKey(&public)

	//fmt.Println(string(Private))
	//fmt.Println(string(publicKey))
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

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func GeneratePublicKeyHash1(publicKey []byte) []byte {
	sha256PubKey := sha256.Sum256(publicKey)
	r := ripemd160.New()
	r.Write(sha256PubKey[:])
	ripPubKey := r.Sum(nil)
	return ripPubKey
}

//func GeneratePublicKeyHash2(publicKey []byte) []byte {
//	sha256PubKey := sha1.Sum(publicKey)
//	r := ripemd160.New()
//	r.Write(sha256PubKey[:])
//	ripPubKey := r.Sum(nil)
//	return ripPubKey
//}

const CHECKSUM_LENGTH = 4

func CheckSumHash1(versionPublickeyHash []byte) []byte {
	versionPublickeyHashSha1 := sha256.Sum256(versionPublickeyHash)
	versionPublickeyHashSha2 := sha256.Sum256(versionPublickeyHashSha1[:])
	tailHash := versionPublickeyHashSha2[:CHECKSUM_LENGTH]
	return tailHash
}

func CheckSumHash2(versionPublickeyHash []byte) []byte {
	versionPublickeyHashSha1 := sha1.Sum(versionPublickeyHash)
	versionPublickeyHashSha2 := sha1.Sum(versionPublickeyHashSha1[:])
	tailHash := versionPublickeyHashSha2[:CHECKSUM_LENGTH]
	return tailHash
}

func Test_acc(t *testing.T) {
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

func Test_sat2(t *testing.T) {
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

	fmt.Println("地址: ", address)
	fmt.Println("地址2: ", address2)
	fmt.Println("私钥: ", wif.String())
}

func Test_sat(t *testing.T) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
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
	fmt.Println("地址: ", address)
	fmt.Println("地址2: ", address2)
	fmt.Println("地址3: ", hash.EncodeAddress())
	fmt.Println("私钥: ", wif.String())
}

func Test_decodedAddress(t *testing.T) {
	decode, bytes, err := bech32.Decode("sat1q37pnjwm7dsdat7zy5z5gq6z7txqmjplvv7v538")
	fmt.Println(decode)
	fmt.Println(err)
	bits, err := bech32.ConvertBits(bytes[1:], 5, 8, true)
	pkhash, err := btcutil.NewAddressPubKeyHash(bits, NetParams)
	address2 := pkhash.EncodeAddress()
	fmt.Println(address2)

}
