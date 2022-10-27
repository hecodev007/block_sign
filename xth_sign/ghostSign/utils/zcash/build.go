package zcash

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/iqoption/zecutil"
)

func BuildRawTx(data []byte, coinNet *chaincfg.Params) (*wire.MsgTx, error) {
	var txInput = new(UtxoParams)
	if err := json.Unmarshal(data, txInput); err != nil {
		return nil, err
	}
	tx := wire.NewMsgTx(4)

	//组装txout输出
	for _, v := range txInput.TxOuts {
		if addr, err := zecutil.DecodeAddress(v.ToAddr, coinNet.Name); err != nil {
			return nil, err
		} else if addrType := CheckAddressType(addr); addrType == -1 {
			return nil, errors.New("unsuport to address: prefix with t1 or t3")
		} else if pkScript, err := zecutil.PayToAddrScript(addr); err != nil {
			fmt.Println("PayToAddrScript 000", err.Error())
			return nil, err
		} else {
			txOut := wire.NewTxOut(v.ToAmount, pkScript)
			tx.AddTxOut(txOut)
		}
	}

	//组装txin输入
	for _, v := range txInput.TxIns {
		//from只支持t1地址
		if addr, err := zecutil.DecodeAddress(v.FromAddr, coinNet.Name); err != nil {
			return nil, err
		} else if addrType := CheckAddressType(addr); addrType != P2PKH {
			return nil, errors.New("unsuport from address: prefix with t1")
		}

		prevTxHash, err := chainhash.NewHashFromStr(v.FromTxid)
		if err != nil {
			return nil, err
		}
		//构造txin输入，注意index的位置配对
		prevOut := wire.NewOutPoint(prevTxHash, v.FromIndex)
		//组装txin模板
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
	}
	//fee := inAmount-outAmount
	return tx, nil
}

type UtxoParams struct {
	TxIns        []UcaTxInTpl  `json:"txIns" binding:"required"` //如果是
	TxOuts       []UcaTxOutTpl `json:"txOuts" binding:"required"`
	ChangeAddr   string        `json:"changeAddr"` //找零地址
	Fee          int64         `json:"fee" binding:"required,min=1000,max=10000000"`
	ExpiryHeight uint32        `json:"expiryHeight"`
}

//utxo模板
type UcaTxInTpl struct {
	FromAddr string `json:"fromAddr"` //来源地址
	//FromPrivkey      string `json:"fromPrivkey"`      //来源地址地址对于的私钥，签名期间赋值
	FromTxid   string `json:"fromTxid"`   //来源UTXO的txid
	FromIndex  uint32 `json:"fromIndex"`  //来源UTXO的txid 地址的下标
	FromAmount int64  `json:"fromAmount"` //来源UTXO的txid 对应的金额
	//FromRedeemScript string `json:"fromRedeemScript"` //多签脚本
}

//输出模板
type UcaTxOutTpl struct {
	ToAddr   string `json:"toAddr"`   //txout地址
	ToAmount int64  `json:"toAmount"` //txout金额
}
