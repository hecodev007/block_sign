package mw

import (
	"encoding/hex"
	"mwSign/common/validator"
)

func BuildTx(params *validator.TelosSignParams) (tx *Transaction, err error) {
	tx = NewTransaction()
	tx.Deadline = params.Deadline
	tx.Timestamp = uint32(params.Deadline)
	tx.AmountNQT = uint64(params.Value.IntPart())
	tx.FeeNQT = uint64(params.Fee.IntPart())

	accountid, err := AddrToAccoutid(params.ToAddress)
	if err != nil {
		return nil, err
	}
	err = tx.SetRecipient(accountid)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func SignTx(tx *Transaction, pri string) (string, error) {
	pub, err := PrivateToPub(pri)
	if err != nil {
		return "", err
	}
	pubytes, _ := hex.DecodeString(pub)
	copy(tx.PublickKey[:], pubytes[0:32])

	rawTx := tx.Seriallize()
	signrawTx, err := Sign(rawTx, pri)

	return signrawTx, err
}
