package types

import (
	// RPC
	"steemsync/utils/rpc/encoding/transaction"

	// Vendor
	"github.com/pkg/errors"
)

// Transaction represents a blockchain transaction.
type Transaction struct {
	RefBlockNum    UInt16     `json:"ref_block_num"`
	RefBlockPrefix UInt32     `json:"ref_block_prefix"`
	Expiration     *Time      `json:"expiration"`
	Operations     Operations `json:"operations"`
	Extensions     []string   `json:"extensions"`
	Signatures     []string   `json:"signatures"`
}

// MarshalTransaction implements transaction.Marshaller interface.
func (tx *Transaction) MarshalTransaction(encoder *transaction.Encoder) error {
	if len(tx.Operations) == 0 {
		return errors.New("no operation specified")
	}

	enc := transaction.NewRollingEncoder(encoder)

	enc.Encode(tx.RefBlockNum)
	enc.Encode(tx.RefBlockPrefix)
	enc.Encode(tx.Expiration)

	enc.EncodeUVarint(uint64(len(tx.Operations)))
	for _, op := range tx.Operations {
		enc.Encode(op)
	}

	// Extensions are not supported yet.
	enc.EncodeUVarint(0)

	return enc.Err()
}

// PushOperation can be used to add an operation into the transaction.
func (tx *Transaction) PushOperation(op Operation) {
	tx.Operations = append(tx.Operations, op)
}

type TransactionInfo struct {
	RefBlockNum    UInt16       `json:"ref_block_num"`
	RefBlockPrefix UInt32       `json:"ref_block_prefix"`
	Expiration     *Time        `json:"expiration"`
	Operations     OperationsV1 `json:"operations"`
	Extensions     []string     `json:"extensions"`
	Signatures     []string     `json:"signatures"`
	TransactionId  string       `json:"transaction_id"`
	BlockNum       int64        `json:"block_num"`
	TransactionNum int64        `json:"transaction_num"`
}

// Operation represents an operation stored in a transaction.
type OperationV1 struct {
	Type  string        `json:"type"`
	Value OperationImpl `json:"value"`
}
type OperationImpl struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Memo   string  `json:"memo"`
	Amount AmountS `json:"amount"`
}

type AmountS struct {
	Amount    string `json:"amount"`
	Precision int    `json:"precision"`
	Nai       string `json:"nai"`
}

type OperationsV1 []OperationV1
