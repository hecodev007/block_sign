package dfn

import (
	"github.com/dfinity/go-dfinity-crypto/bls"
	"testing"
)

func Test_acc(t *testing.T){
	var sec bls.SecretKey
	sec.SetByCSPRNG()
	t.Log("sec:", sec.GetHexString())
	t.Log("create public key")
	pub := sec.GetPublicKey()
	t.Log("pub:", pub.GetHexString())
	sign := sec.Sign("123456")
	t.Log("sign:", sign.GetHexString())
}
