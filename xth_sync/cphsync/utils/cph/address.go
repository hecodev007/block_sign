package cph

import (
	"cphsync/utils/sha3"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"

	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var addr_prefix = "CPH"

func ToAddressCypherium(addrhex string) (addr string) {
	addr = strings.Replace(strings.ToLower(addrhex), "0x", "cph", 1)
	return ToCphAddress(addr)
}

func ToCphAddress(addr string) (address string) {
	comaddr := ToCommonAddress(addr)
	return strings.Replace(comaddr.String(), "0x", "CPH", 1)
}

func PubKeyToAddressCypherium(publichex string) (addr string, err error) {
	publickbytes, err := hex.DecodeString(publichex)
	if err != nil {
		return
	}
	addrSha := sha3.Sum256(publickbytes)
	addr160 := ripemd160.New()
	addr160.Write(addrSha[:])

	address := common.BytesToAddress(addr160.Sum(nil))
	return strings.Replace(address.String(), "0x", "cph", 1), nil
}
