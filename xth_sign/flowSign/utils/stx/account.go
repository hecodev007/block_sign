package stx

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"math/big"
	"errors"
	"strings"
)
var (
secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
)
func GentAccount() (addr string,private string,err error) {
	start:
	pri,err :=ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)

	pub :=pri.PublicKey
	pubytes :=elliptic.Marshal(secp256k1.S256(),pub.X,pub.Y)
	pribytes :=math.PaddedBigBytes(pri.D, pri.Params().BitSize/8)

	private = hex.EncodeToString(pribytes)
	if strings.HasPrefix(private,"fffffffffffffffffffffffffffffff"){
		goto start
	}
	addr,err = PubToAddr(pubytes)
	return
}
func HexToPri(hexkey string)(*ecdsa.PrivateKey, error){
	b, err := hex.DecodeString(hexkey)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}
	return toECDSA(b,true)
}
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}
func PriToAddr(hexPri string)(addr string,err error){
	pri,err :=HexToPri(hexPri)
	if err != nil {
		return "",err
	}
	pub :=pri.PublicKey
	pubytes :=elliptic.Marshal(secp256k1.S256(),pub.X,pub.Y)
	return PubToAddr(pubytes)
}
func PubToAddr(pub []byte)(addr string,err error){
	sum1 :=sha256.Sum256(pub)
	digest := ripemd160.New()
	digest.Write(sum1[:])
	rip160 :=digest.Sum(nil)
	return C32_check_encode(22,rip160)
}