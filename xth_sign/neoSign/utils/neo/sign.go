package neo

import (
	"errors"
	"github.com/joeqian10/neo-gogogo/helper"
	"neoSign/common/validator"

	"github.com/o3labs/neo-utils/neoutils"
	"github.com/o3labs/neo-utils/neoutils/smartcontract"
	"github.com/shopspring/decimal"
	 "github.com/joeqian10/neo-gogogo/tx"
	"github.com/joeqian10/neo-gogogo/wallet/keys"
)
func BuildClaim(params *validator.SignParams)(ctx *tx.ClaimTransaction,err error){
	var claims []*tx.CoinReference
	for _,txin := range params.TxIns{
		txiduint160,err :=helper.UInt256FromString(txin.FromTxid)
		if err != nil{
			return nil,err
		}
		claim := &tx.CoinReference{
			PrevHash:txiduint160,
			PrevIndex:uint16(txin.FromIndex),
		}
		claims = append(claims,claim)
	}
	ctx = tx.NewClaimTransaction(claims)
	gasToken, _ := helper.UInt256FromString(tx.GasTokenId)
	amount := helper.Fixed8{
		Value:params.TxOuts[0].ToAmountInt64,
	}
	caddr,err :=helper.AddressToScriptHash(params.TxOuts[0].ToAddr)
	if err != nil{
		return
	}
	output := tx.NewTransactionOutput(gasToken, amount, caddr)
	var outputs []*tx.TransactionOutput

	outputs = append(outputs, output)
	ctx.Outputs = outputs
	return
}
func SignClaim(ctx *tx.ClaimTransaction,pris []string)(rawtx string,txhash string,err error){
	for _,pri := range pris{
		key,err := keys.NewKeyPairFromWIF(pri)
		if err != nil{
			return "","",err
		}
		if err = tx.AddSignature(ctx,key);err != nil {
			return "","",err
		}
	}
	return ctx.RawTransactionString(),ctx.HashString(),nil
}
func BuildTx(params *validator.SignParams) (tx smartcontract.Transaction, err error) {
	tx = smartcontract.NewInvocationTransactionPayable()
	var inMount int64
	inBuilder := smartcontract.NewScriptBuilder()
	inBuilder.PushLength(len(params.TxIns))
	for _, in := range params.TxIns {
		inMount += in.FromAmountInt64
		value, _ := decimal.NewFromInt(in.FromAmountInt64).Shift(-8).Float64()
		utxo := smartcontract.UTXO{
			Index: in.FromIndex,
			TXID:  in.FromTxid,
			Value: value,
		}
		inBuilder.Push(utxo)
	}
	tx.Inputs = inBuilder.ToBytes()

	outBuilder := smartcontract.NewScriptBuilder()
	outBuilder.PushLength(len(params.TxOuts))
	var outAmount int64
	for _, out := range params.TxOuts {
		outAmount += out.ToAmountInt64

		to := smartcontract.ParseNEOAddress(out.ToAddr)
		txout := smartcontract.TransactionOutput{
			Asset:   smartcontract.NEO,
			Address: to,
			Value:   out.ToAmountInt64,
		}
		outBuilder.Push(txout)
	}
	tx.Outputs = outBuilder.ToBytes()

	if inMount < outAmount {
		return tx, errors.New("insufient value")
	}
	if inMount > outAmount+10000000 {
		return tx, errors.New("in.amount > out.amount+0.1")
	}

	return tx, nil
}

func Sign(tx smartcontract.Transaction, private string) (rawTx string, txid string, err error) {
	wallet, err := neoutils.GenerateFromWIF(private)
	if err != nil {
		return "", "", err
	}
	signedData, err := neoutils.Sign(tx.ToBytes(), neoutils.BytesToHex(wallet.PrivateKey))
	if err != nil {
		return "", "", err
	}
	signature := smartcontract.TransactionSignature{
		SignedData: signedData,
		PublicKey:  wallet.PublicKey,
	}
	scripts := []interface{}{signature}
	txScripts := smartcontract.NewScriptBuilder().GenerateVerificationScripts(scripts)
	tx.Script = txScripts
	tx.GAS = uint64(490)
	tx.Data = []byte{0x12, 0x34}
	tx.Attributes = []byte{0x12, 0x34}
	return neoutils.BytesToHex(tx.ToBytes()), tx.ToTXID(), nil
}
