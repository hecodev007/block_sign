package stx

import "flowSign/common/validator"
//const Chainid = uint32(0x00000001)
//const TxVersion =  uint8(0x00)
const Chainid = uint32(0x80000000)
const TxVersion =  uint8(0x80)
func BuildTx(params *validator.SignParams)(tx *Transaction,err error){
	tx = &Transaction{
		Version:           TxVersion,
		Chainid:           Chainid,
		Auth:              SinglesigSpendingCondition{
			Type:0x04,
			Nonce:params.Nonce,
			TxFee: params.Fee,
			HashModel: 0x00,
			//Signer [20]byte
			KeyEncoding:0x01,
			//Sigature [65]byte
		},
		AuthModel:         0x03,
		PostConditionMode: 0x01,
		PostConditions:    0,
		Payload:           TokenTransfer{
			Receipient:StandardPrincipalData{
				Type:0x05,
				//to.version
				//to.bytes
			},
			Amount: params.Value,
			//memo:[34]byte
		},
	}
	_,fromBytes,err := C32_check_decode(params.FromAddress)
	if err != nil {
		return nil,err
	}
	copy(tx.Auth.Signer[:],fromBytes)
	toVersion,toBytes,err := C32_check_decode(params.ToAddress)
	if err != nil{
		return nil,err
	}
	tx.Payload.Receipient.Version = toVersion
	copy(tx.Payload.Receipient.Bytes[:],toBytes)
	copy(tx.Payload.Memo[:],params.Memo)
	return tx,nil
}