package transactions

import (
	"github.com/JFJun/helium-go/keypair"
	"github.com/JFJun/helium-go/protos"
	"github.com/golang/protobuf/proto"
)

type PaymentV1Tx struct {
	Payer  *keypair.Addressable
	Payee  *keypair.Addressable
	Amount uint64
	Fee    uint64
	Nonce  uint64
	Sig    []byte
}

func NewPaymentV1Tx(from, to *keypair.Addressable, amount, fee, nonce uint64, sig []byte) *PaymentV1Tx {
	return &PaymentV1Tx{
		Payer:  from,
		Payee:  to,
		Amount: amount,
		Fee:    fee,
		Nonce:  nonce,
		Sig:    sig,
	}
}

func (v1 *PaymentV1Tx) SetFee(fee uint64) {
	v1.Fee = fee
}

func (v1 *PaymentV1Tx) BuildTransaction(isForSign bool) ([]byte, error) {
	btpV1 := new(protos.BlockchainTxnPaymentV1)
	if v1.Payer != nil {
		btpV1.Payer = v1.Payer.GetBin()
	}
	if v1.Payee != nil {
		btpV1.Payee = v1.Payee.GetBin()
	}
	btpV1.Amount = v1.Amount
	if v1.Fee > 0 {
		btpV1.Fee = v1.Fee
	}
	btpV1.Nonce = v1.Nonce
	if v1.Sig != nil && !isForSign {
		btpV1.Signature = v1.Sig
	}

	return proto.Marshal(btpV1)
}
func (v1 *PaymentV1Tx) Serialize() ([]byte, error) {
	txn := new(protos.BlockchainTxn)
	var btpV1 protos.BlockchainTxnPaymentV1
	data, err := v1.BuildTransaction(false)
	err = proto.Unmarshal(data, &btpV1)
	if err != nil {
		return nil, err
	}
	bp := protos.BlockchainTxn_Payment{Payment: &btpV1}
	txn.Txn = &bp
	return proto.Marshal(txn)
}
func (v1 *PaymentV1Tx) SetSignature(sig []byte) {
	v1.Sig = sig
	return
}
