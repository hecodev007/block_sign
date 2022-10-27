package txbuilder

import (
	"btmSign/bytom/crypto/ed25519/chainkd"
	chainjson "btmSign/bytom/encoding/json"
	"btmSign/bytom/protocol/bc/types"
)

func BuildSignatureDataToTplByJun(tpl *Template, sig, data []byte, i int) *Template {
	sigInst := tpl.SigningInstructions[i]
	if len(sigInst.WitnessComponents) == 0 {
		witCom := []witnessComponent{
			&RawTxSigWitness{
				Quorum: 1,
				Sigs:   []chainjson.HexBytes{sig},
			},
			DataWitness(data),
		}
		sigInst.WitnessComponents = witCom
	}

	//err := materializeWitnesses(tpl)
	//fmt.Println(err)
	return tpl

}

func CheckTpl(tpl *Template) (*Template, error) {
	if err := materializeWitnesses(tpl); err != nil {
		return nil, err
	}

	//if !testutil.DeepEqual(tx, tpl.Transaction) {
	//	return nil,errors.New(fmt.Sprintf("tx:%v result is equal to want:%v", tx, tpl.Transaction))
	//}
	return tpl, nil
}

func BuildTransaction(tx *types.Tx) *Template {
	tpl := &Template{}
	tpl.AllowAdditional = false
	for i, _ := range tx.Inputs {
		instruction := &SigningInstruction{}
		instruction.Position = uint32(i)
		// Empty signature arrays should be serialized as empty arrays, not null.
		if instruction.WitnessComponents == nil {
			instruction.WitnessComponents = []witnessComponent{}
		}
		tpl.SigningInstructions = append(tpl.SigningInstructions, instruction)
	}
	tpl.Transaction = tx
	tpl.Fee = CalculateTxFee(tpl.Transaction) //计算手续费
	return tpl
}

type InputAndSigInst struct {
	Input   *types.TxInput
	SigInst *SigningInstruction
}

func BuildTx2(inputs []InputAndSigInst, outputs []*types.TxOutput) (*Template, *types.TxData, error) {
	tpl := &Template{}
	tx := &types.TxData{}
	// Add all the built outputs.
	tx.Outputs = append(tx.Outputs, outputs...)

	// Add all the built inputs and their corresponding signing instructions.
	for _, in := range inputs {
		// Empty signature arrays should be serialized as empty arrays, not null.
		in.SigInst.Position = uint32(len(inputs) - 1)
		if in.SigInst.WitnessComponents == nil {
			in.SigInst.WitnessComponents = []witnessComponent{}
		}
		tpl.SigningInstructions = append(tpl.SigningInstructions, in.SigInst)
		tx.Inputs = append(tx.Inputs, in.Input)
	}

	tpl.Transaction = types.NewTx(*tx)
	return tpl, tx, nil
}

func Sign2(tpl *Template, xprv chainkd.XPrv) error {
	for i, sigInst := range tpl.SigningInstructions {
		h := tpl.Hash(uint32(i)).Byte32()
		sig := xprv.Sign(h[:])
		rawTxSig := &RawTxSigWitness{
			Quorum: 1,
			Sigs:   []chainjson.HexBytes{sig},
		}
		sigInst.WitnessComponents = append([]witnessComponent{rawTxSig}, sigInst.WitnessComponents...)
	}
	return materializeWitnesses(tpl)
}
