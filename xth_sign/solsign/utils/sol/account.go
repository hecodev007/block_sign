package sol

import (
	"encoding/hex"

	"github.com/portto/solana-go-sdk/types"
)

func GenAccount() (addr string, pri string, err error) {
	acc := types.NewAccount()
	seed := acc.PrivateKey.Seed()
	addr = acc.PublicKey.ToBase58()
	pri = hex.EncodeToString(seed)
	return
}
