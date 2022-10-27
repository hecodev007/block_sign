package ecash

import (
	"fmt"
	"xecsign/common/log"
	"xecsign/common/validator"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil"

	"errors"
	"strings"
)

func init() {
	chaincfg.MainNetParams.CashAddressPrefix = "ecash"
}

func BuildTx(params *validator.SignParams) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.Outs {
		_, pubkeyscript, err := CreatePayScript(out.ToAddr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.ToAmountInt64, pubkeyscript))
		outMount += out.ToAmountInt64
	}
	for _, in := range params.Ins {
		txhash, err := chainhash.NewHashFromStr(in.FromTxid)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(txhash, in.FromIndex)
		txIn := wire.NewTxIn(prevOut, nil)
		tx.AddTxIn(txIn)
		inMount += in.FromAmountInt64
	}
	log.Info(inMount)
	//额度是否足够
	if inMount < outMount {
		return nil, errors.New("insufficient mount")
	}
	//0.1 dash fee
	if inMount > outMount+100000 {
		return nil, errors.New("too many tx.fee")
	}
	return tx, nil
}
func SignTx(tx *wire.MsgTx, index int, amount int64, pri string) (*wire.MsgTx, error) {
	wif, err := bchutil.DecodeWIF(pri)
	if err != nil {
		return tx, err
	}
	pk := (*bchec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := bchutil.NewAddressPubKeyHash(bchutil.Hash160(pk), &chaincfg.MainNetParams)
	if err != nil {
		return tx, err
	}
	from_cashaddr := pkhash.EncodeAddress()
	_, pkScript, err := CreatePayScript(from_cashaddr)
	if err != nil {
		return tx, err
	}
	sigScript, err := txscript.SignatureScript(tx, index, amount, pkScript, txscript.SigHashAll, wif.PrivKey, true)
	if err != nil {
		return tx, err
	}
	tx.TxIn[index].SignatureScript = sigScript
	return tx, nil
}
func checkTplAddr(addr string) error {
	if strings.HasPrefix(addr, "ecash:") ||
		strings.HasPrefix(addr, "q") ||
		strings.HasPrefix(addr, "p") ||
		strings.HasPrefix(addr, "1") ||
		strings.HasPrefix(addr, "3") {
		return nil
	}
	return fmt.Errorf("Unsupported  out address type,address:%s", addr)
}

//创建地址交易脚本
func CreatePayScript(addrStr string) (bchutil.Address, []byte, error) {
	if err := checkTplAddr(addrStr); err != nil {
		return nil, nil, err
	}

	//ecash
	addr, err := bchutil.DecodeAddress(addrStr, &chaincfg.MainNetParams)
	if err != nil {
		//bitcoincash
		bitcoincashParams := chaincfg.MainNetParams
		bitcoincashParams.CashAddressPrefix = "bitcoincash"
		addr, err = bchutil.DecodeAddress(addrStr, &bitcoincashParams)
		if err != nil {
			return nil, nil, fmt.Errorf("DecodeAddress error:%s", err.Error())
		}
	}

	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}
