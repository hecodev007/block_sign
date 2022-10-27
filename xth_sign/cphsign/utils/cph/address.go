package cph

import (
	"cphsign/utils/sha3"
	crand "crypto/rand"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/ed25519"

	"golang.org/x/crypto/ripemd160"

	"github.com/ethereum/go-ethereum/common"
)

var addr_prefix = "CPH"

//
//func GenAccount() (addr string, prihex string, err error) {
//	key, err := crypto.GenerateKey()
//	if err != nil {
//		return
//	}
//	address := crypto.PubkeyToAddress(key.PublicKey)
//	addr = strings.Replace(address.String(), "0x", addr_prefix, 1)
//	prihex = hex.EncodeToString(crypto.FromECDSA(key))
//	return
//}
func GenAccount() (addr string, prihex string, err error) {
	pub, pri, err := ed25519.GenerateKey(crand.Reader)
	if err != nil {
		return
	}
	address := PubKeyToAddressCypherium(pub)
	addr = strings.Replace(address.String(), "0x", addr_prefix, 1)
	pri.Seed()
	prihex = hex.EncodeToString(pri.Seed())
	return
}

func PubKeyToAddressCypherium(publicKey []byte) common.Address {
	addrSha := sha3.Sum256(publicKey)
	addr160 := ripemd160.New()
	addr160.Write(addrSha[:])

	return common.BytesToAddress(addr160.Sum(nil))
}
