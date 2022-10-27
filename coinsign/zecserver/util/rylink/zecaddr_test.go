package rylink

import (
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"testing"
)

func TestCreateAddress2(t *testing.T) {
	t.Log(CreateAddress2())
}

func TestCreateAddress(t *testing.T) {
	decoded := base58.Decode("t3N48xQxTuFUDqDsL7h9UBUmNQg74jtkViC")
	var addr *btcutil.AddressPubKeyHash
	var err error
	if addr, err = btcutil.NewAddressPubKeyHash(decoded[2:len(decoded)-4], netParams); err != nil {
		t.Fatal(err)
	}
	t.Log(addr.String())
}
