package crust

import (
	"github.com/JFJun/substrate-go/tx"
)

func CreateTransaction(from, to string, amount, nonce, fee uint64, genesisHash, blockHash string, blockNumber uint64, specVersion, transactionVersion uint32, callId string) *tx.Transaction {
	tx2 := tx.CreateTransaction(from, to, amount, nonce, uint64(0))
	tx2.SetGenesisHashAndBlockHash(genesisHash, blockHash, blockNumber)
	tx2.SetSpecVersionAndCallId(specVersion, transactionVersion, callId)
	return tx2
}

func SignTransaction(tx2 *tx.Transaction, private string) (rawTx string, err error) {
	_, message, err := tx2.CreateEmptyTransactionAndMessage()
	if err != nil {
		return "", err
	}
	sig, err := tx2.SignTransaction(private, message)
	if err != nil {
		return "", err
	}
	tt, err := tx2.GetSignTransaction(sig)
	if err != nil {
		return "", err
	}
	return tt, nil
}
