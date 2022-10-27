package biw

import (
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/ed25519"
	"github.com/algorand/go-algorand-sdk/types"
	"biwSign/common/validator"
)
func GenAccount() (address string, private string, err error) {
	// Generate an ed25519 keypair. This should never fail
	pk, pri, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	// Convert the public key to an address
	var a types.Address
	n := copy(a[:], pk)
	if n != ed25519.PublicKeySize {
		return "","",errors.New("generated public key is the wrong size")
	}

	return a.String(),hex.EncodeToString(pri),nil
}

func SignTx(params *validator.SignParams,pri string) (txid,rawTx string,err error) {
	return "", "", nil
}
