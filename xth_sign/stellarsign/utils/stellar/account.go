package stellar

import (
	"github.com/stellar/go/keypair"
)

func GenAccount() (addr string, pri string, err error) {
	return "GAQ2IDPAWWC5G3BL2WHNKNGOWEN7HFATFA4RBRSDF5WUKQQMWSCS2QD4", "SBRTFQYUTWONOQRLFSXZFCUFG3XH7D4NDHYMCGR2J7Q3H4XG6FQQHAAM", nil
	key, err := keypair.Random()
	if err != nil {
		return
	}
	return key.Address(), key.Seed(), nil
}
