package algo

import (
	"encoding/hex"

	"github.com/algorand/go-algorand-sdk/crypto"
	//"github.com/algorand/go-algorand-sdk/types"
)

func GenAccount() (addr string, pri string, err error) {
	acc := crypto.GenerateAccount()
	//pub := types.Address{}
	//copy(pub[:], acc.PublicKey[:])
	//println(pub.String())
	return acc.Address.String(), hex.EncodeToString(acc.PrivateKey[:]), nil
}
