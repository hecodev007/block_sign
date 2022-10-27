package stellar

import (
	"testing"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"
)

func Test_acc(t *testing.T) {
	key, err := keypair.Random()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(key.Address(), key.Seed())
	//acc_test.go:15: GAQ2IDPAWWC5G3BL2WHNKNGOWEN7HFATFA4RBRSDF5WUKQQMWSCS2QD4 SBRTFQYUTWONOQRLFSXZFCUFG3XH7D4NDHYMCGR2J7Q3H4XG6FQQHAAM

	seedBytes, err := strkey.Decode(strkey.VersionByteSeed, key.Seed())
	if err != nil {
		t.Fatal(err.Error())
	}
	var seed2 [32]byte
	copy(seed2[:], seedBytes)
	key, err = keypair.FromRawSeed(seed2)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(key.Address(), key.Seed())
}
