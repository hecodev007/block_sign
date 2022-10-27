package zcash

import (
	"encoding/hex"
	"errors"
	"strings"
	"zenSign/common"
	"zenSign/common/log"

	"github.com/btcsuite/btcd/txscript"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/iqoption/zecutil"
)

func BuildRawTx(params *common.ZenSignParams, coinNet *chaincfg.Params) (*wire.MsgTx, error) {
	txInput := params.Data

	tx := wire.NewMsgTx(1)
	var inAmount, outAmount int64
	//组装txout输出
	for _, v := range txInput.TxOuts {
		if pkScript, err := PayToAddrScript(v.ToAddr, params.Data.BlockHash, params.Data.BlockHeight); err != nil {
			log.Info("PayToAddrScript 000", err.Error())
			return nil, err
		} else {
			txOut := wire.NewTxOut(v.ToAmount, pkScript)
			tx.AddTxOut(txOut)
		}
		outAmount += v.ToAmount
	}

	//组装txin输入
	for _, v := range txInput.TxIns {
		inAmount += v.FromAmount
		//from只支持t1地址
		if addr, err := zecutil.DecodeAddress(v.FromAddr, coinNet.Name); err != nil {
			return nil, err
		} else if addrType := CheckAddressType(addr); addrType != P2PKH {
			return nil, errors.New("unsuport from address: prefix with t1")
		}

		prevTxHash, err := chainhash.NewHashFromStr(v.FromTxid)
		if err != nil {
			log.Info(err.Error())
			return nil, err
		}
		//构造txin输入，注意index的位置配对
		prevOut := wire.NewOutPoint(prevTxHash, v.FromIndex)
		//组装txin模板
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
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
func PayToAddrScript(addr string, blockHash string, blockHeight int64) ([]byte, error) {
	address, err := zecutil.DecodeAddress(addr, chaincfgParams.Name)
	if err != nil {
		return nil, err
	}
	hash, err := hex.DecodeString(blockHash)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(addr, "zs") {
		return txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(address.ScriptAddress()).
			AddOp(txscript.OP_EQUAL).AddData(reverse(hash)).AddInt64(blockHeight).AddOp(txscript.OP_NOP5).Script()
	} else {
		return txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
			AddData(address.ScriptAddress()).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
			AddData(reverse(hash)).AddInt64(blockHeight).AddOp(txscript.OP_NOP5).Script()
	}

}
func reverse(arr []byte) []byte {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		temp := arr[length-1-i]
		arr[length-1-i] = arr[i]
		arr[i] = temp
	}
	return arr
}
