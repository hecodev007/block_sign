package atom

import (
	"atomSign/common/validator"
	"encoding/hex"
)

func GenAccount() (address string, private string, err error) {
	acc := GenerateAccount()

	return acc.Address.String(), string(acc.PrivateKey[:]), nil
}

func SignTx(params *validator.SignParams, pri string) (rawTx string, err error) {
	signer := CreateSigner()
	rawTx, err = signer.SignTx(params, pri)
	return
}
func ToACT(pri string) (address string, private string, err error) {

	prikey, err := hex.DecodeString(pri)
	if err != nil {
		panic(err.Error())
	}
	act, err := MakeAccount(prikey)
	if err != nil {
		panic(err.Error())
	}
	return act.Address.String(), string(act.PrivateKey[:]), nil
}
