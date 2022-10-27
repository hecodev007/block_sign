package xlm

import (
	"github.com/stellar/go/keypair"
)

func GenAccount() (addr string, private string, err error) {
	full, err := keypair.Random()
	if err != nil {
		return "", "", err
	}
	return full.Address(), full.Seed(), nil
}
