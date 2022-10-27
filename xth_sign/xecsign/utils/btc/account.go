package btc

import (
	"xecsign/common/validator"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

var NetParams *chaincfg.Params

func init() {
	NetParams = new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x00
	NetParams.ScriptHashAddrID = 0x05
	NetParams.PrivateKeyID = 0x80
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
}

func GenAccount() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := wif.SerializePubKey()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	if err != nil {
		return "", "", err
	}
	address = pkhash.EncodeAddress()
	return address, wif.String(), nil
}

func BuildTx(params *validator.SignParams) (tx *wire.MsgTx, err error) {
	tx = wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.Outs {
		if _, err := btcutil.DecodeAddress(out.ToAddr, NetParams); err != nil {
			return nil, err
		}
		outaddr, err := btcutil.DecodeAddress(out.ToAddr, NetParams)
		if err != nil {
			return nil, err
		}
		pubkeyscript, err := txscript.PayToAddrScript(outaddr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.ToAmountInt64, pubkeyscript))
		outMount += out.ToAmountInt64
	}
	for _, in := range params.Ins {
		if _, err := btcutil.DecodeAddress(in.FromAddr, NetParams); err != nil {
			return nil, err
		}
		//txhash, err := wire.NewShaHashFromStr(in.FromTxid)
		txhash, err := chainhash.NewHashFromStr(in.FromTxid)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(txhash, in.FromIndex)
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
		inMount += in.FromAmountInt64
	}
	//额度是否足够
	if inMount < outMount+100000 {
		return nil, errors.New("insufficient mount or fee(0.001)")
	}
	//max 1 fee
	if inMount > outMount+100000000 {
		return nil, errors.New("too many tx.fee")
	}
	return tx, nil
}
