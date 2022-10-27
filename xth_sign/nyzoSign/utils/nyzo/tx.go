package nyzo

import (
	"nyzoSign/common/validator"
	//"github.com/onethefour/go_nyzo/nyzo/blockchain_data"
	"github.com/onethefour/go_nyzo/nyzo/blockchain_data"
	"github.com/onethefour/go_nyzo/pkg/identity"

	"time"
)

func BuildTx(params *validator.TelosSignParams) (tx *blockchain_data.Transaction,err error){
	tx = new(blockchain_data.Transaction)
	tx.Type = blockchain_data.TransactionTypeStandard
	tx.Amount = params.Value.IntPart()
	tx.Timestamp = time.Now().Unix()
	tx.SenderId ,err = identity.FromNyzoString(params.FromAddress)
	if err != nil {
		return
	}
	tx.RecipientId,err = identity.FromNyzoString(params.ToAddress)
	if err != nil {
		return
	}
	return
}

func SignTx(tx *blockchain_data.Transaction,pri string)(rawTx []byte,err error){
	priBytes,err := identity.FromNyzoString(pri)
	if err != nil {
		return
	}
	acc,err := identity.FromPrivateKey(priBytes)
	if err != nil {
		return
	}
	txBytes := tx.Serialize(true)
	signratrue := acc.Sign(txBytes)
	tx.Signature = signratrue
	return tx.ToBytes(),nil
}