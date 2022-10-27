package atom2

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenAccount() (address string, private string, err error) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	return acc.String(), string(privKey.Key[:]), nil
}
