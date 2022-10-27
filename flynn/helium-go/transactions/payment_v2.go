package transactions

import (
	"github.com/JFJun/helium-go/keypair"
	"github.com/JFJun/helium-go/protos"
	"github.com/golang/protobuf/proto"
)

type PaymentV2Tx struct {
	Payer    *keypair.Addressable
	Fee      uint64
	Payments []*Payment
	Nonce    uint64
	Sig      []byte
}

type Payment struct {
	Payee  *keypair.Addressable
	Amount uint64
}

func NewPaymentV2Tx(from *keypair.Addressable, toAmount map[string]uint64, fee, nonce uint64, sig []byte) *PaymentV2Tx {
	if toAmount == nil {
		return nil
	}
	var payments []*Payment
	for to, amount := range toAmount {
		payee := keypair.NewAddressable(to)
		if payee == nil {
			return nil
		}
		payment := &Payment{
			Payee:  payee,
			Amount: amount,
		}
		payments = append(payments, payment)
	}
	v2 := &PaymentV2Tx{
		Payer: from,
		Fee:   fee,
		Nonce: nonce,
		Sig:   sig,
	}
	v2.Payments = payments
	return v2
}

func (v2 *PaymentV2Tx) SetFee(fee uint64) {
	v2.Fee = fee
}

func (v2 *PaymentV2Tx) BuildTransaction(isForSign bool) ([]byte, error) {
	btpV2 := new(protos.BlockchainTxnPaymentV2)
	if v2.Payer != nil {
		btpV2.Payer = v2.Payer.GetBin()
	}
	//if v1.Payee!=nil {
	//	btpV1.Payee = v1.Payee.GetBin()
	//}
	if v2.Fee > 0 {
		btpV2.Fee = v2.Fee
	}
	if v2.Sig != nil && !isForSign {
		btpV2.Signature = v2.Sig
	}
	//btpV1.Amount = v1.Amount
	btpV2.Nonce = v2.Nonce
	var payments []*protos.Payment
	for _, payment := range v2.Payments {
		p := new(protos.Payment)
		p.Payee = payment.Payee.GetBin()
		p.Amount = payment.Amount
		payments = append(payments, p)
	}
	btpV2.Payments = payments
	return proto.Marshal(btpV2)
}
func (v2 *PaymentV2Tx) Serialize() ([]byte, error) {
	txn := new(protos.BlockchainTxn)
	data, err := v2.BuildTransaction(false)
	if err != nil {
		return nil, err
	}
	var btpV2 protos.BlockchainTxnPaymentV2
	err = proto.Unmarshal(data, &btpV2)
	if err != nil {
		return nil, err
	}
	bp := protos.BlockchainTxn_PaymentV2{PaymentV2: &btpV2}
	txn.Txn = &bp
	return proto.Marshal(txn)
}
func (v2 *PaymentV2Tx) SetSignature(sig []byte) {
	v2.Sig = sig
	return
}
