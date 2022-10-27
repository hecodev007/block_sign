package fil

import (
	"encoding/hex"
	"github.com/filecoin-project/go-address"
	"testing"
)

func Test_address(t *testing.T) {
	pri, err := new(SecpSigner).GenPrivate()
	if err != nil {
		t.Fatal(err.Error())
	}
	pub, err := new(SecpSigner).ToPublic(pri)
	if err != nil {
		t.Fatal(err.Error())
	}
	addr, err := address.NewSecp256k1Address(pub)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr.String(), hex.EncodeToString(pri))
	address.CurrentNetwork = address.Testnet
	t.Log(addr.String(), hex.EncodeToString(pri))

}
