package nyzo

import (
	"encoding/hex"
	//"github.com/cryptic-monk/go-nyzo/pkg/identity"
	"testing"
	//"github.com/Open-Nyzo/go-nyzo/pkg/identity"
	"github.com/onethefour/go_nyzo/pkg/identity"
	"github.com/qqvv/go-nyzo/crypto"
)

func Test_acc(t *testing.T) {
	t.Log(GenAccount())
}
func Test_msg(t *testing.T) {
	pri := crypto.GenPrivKey()
	t.Log(pri.String())
	acc, err := identity.FromPrivateKey(pri[0:32])
	if err != nil {
		panic(err.Error())
	}
	t.Log(hex.EncodeToString(acc.PublicKey[:]))
	t.Log(acc.NyzoStringPrivate, acc.NyzoStringPublic)
	pub, err := identity.FromNyzoString(acc.NyzoStringPublic)
	if err != nil {
		panic(err.Error())
	}
	t.Log(hex.EncodeToString(pub))

	priHex, err := identity.FromNyzoString(acc.NyzoStringPrivate)
	if err != nil {
		panic(err.Error())
	}
	t.Log(hex.EncodeToString(priHex))
}
func Test_valid(t *testing.T) {
	t.Log(ValidAddress("id__86NfwfQXH6ST0jhVmrvkHmPrBD~XJ9EjG6Geb~U~vSbEkVTvpJ8"))
}
