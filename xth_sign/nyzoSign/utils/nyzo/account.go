package nyzo

import (
	"crypto/rand"
	"strings"

	"github.com/onethefour/go_nyzo/pkg/identity"
)

func GenAccount() (addr string, pri string, err error) {
	priBytes := RandBytes(32)
	acc, err := identity.FromPrivateKey(priBytes)
	if err != nil {
		return "", "", err
	}

	return acc.NyzoStringPublic, acc.NyzoStringPrivate, nil
}
func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func ValidAddress(addr string) (ok bool, msg string) {
	defer func() {
		if err := recover(); err != nil {
			ok = false
			msg = "校验地址" + addr + " 失败"
		}
	}()

	if !strings.HasPrefix(addr, "id__") {
		return false, "校验地址" + addr + " 前缀不对:id__"
	}

	addrBytes, err := identity.FromNyzoString(addr)
	if err != nil {
		return false, err.Error()
	}
	addr2 := identity.ToNyzoString(2, addrBytes)
	if addr != addr2 {
		return false, "校验地址" + addr + " 失败"
	}
	return true, ""
}
