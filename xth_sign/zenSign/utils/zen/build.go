package zen

import (
	"encoding/hex"
	"errors"
	"zenSign/common"

	"github.com/HorizenOfficial/rosetta-zen/zend/txscript"
	"github.com/HorizenOfficial/rosetta-zen/zenutil"

	"github.com/HorizenOfficial/rosetta-zen/zend/chaincfg"
	"github.com/HorizenOfficial/rosetta-zen/zend/chaincfg/chainhash"
	"github.com/HorizenOfficial/rosetta-zen/zend/wire"
)

func BuildRawTx(params *common.ZenSignParams) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	var inAmount, outAmount int64
	for _, in := range params.TxIns {
		txid, err := chainhash.NewHashFromStr(in.FromTxid)
		if err != nil {
			return nil, err
		}
		pkscript, err := hex.DecodeString(in.FromScript)
		if err != nil {
			return nil, err
		}
		txpt := wire.NewOutPoint(txid, in.FromIndex)
		txin := wire.NewTxIn(txpt, pkscript)
		tx.AddTxIn(txin)
		inAmount += in.FromAmount
	}

	for _, out := range params.TxOuts {
		toaddress, err := zenutil.DecodeAddress(out.ToAddr, &chaincfg.MainNetParams)
		if err != nil {
			return nil, err
		}
		blockhash, err := hex.DecodeString(params.BlockHash)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrReplayOutScript(toaddress, blockhash, params.BlockHeight)
		if err != nil {
			return nil, err
		}
		txOut := wire.NewTxOut(out.ToAmount, pkScript)
		tx.AddTxOut(txOut)
		outAmount += out.ToAmount
	}
	fee := inAmount - outAmount
	if fee < 0 {
		return nil, errors.New("insuffient balance")
	}
	if fee > 1000000 {
		return nil, errors.New("too many tx.fee")
	}
	return tx, nil
}
