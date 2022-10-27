package dogeutil

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/shopspring/decimal"
	"log"
	"strings"
)

//该版本暂时只支持常规地址版本发送
const VERSION = int32(1) //版本定制为2

func SignTxTpl(tpl *DogeTxTpl) (string, error) {
	if len(tpl.TxIns) < 1 || len(tpl.TxOuts) < 1 {
		return "", errors.New("error input data")
	}

	redeemTx := wire.NewMsgTx(VERSION)

	//组装txout输出
	for _, v := range tpl.TxOuts {
		toAmInt64 := decimal.NewFromInt(v.ToAmountInt64)

		//A为多签地址，暂时不支持
		if !strings.HasPrefix(v.ToAddr, "D") {
			return "", fmt.Errorf("Unsupported  out address type,address:%s", v.ToAddr)
		}
		_, toPkScript, err := createPayScript(v.ToAddr, coinNetPrams)
		if err != nil {
			return "", err
		}
		//构造txout输出，注意是否存在找零
		txOut := wire.NewTxOut(toAmInt64.IntPart(), toPkScript)
		redeemTx.AddTxOut(txOut)
	}
	//组装txin输入
	for _, v := range tpl.TxIns {
		prevTxHash, err := chainhash.NewHashFromStr(v.FromTxid)
		if err != nil {
			return "", fmt.Errorf("get prevTxHash error:%v", err)
		}
		//构造txin输入，注意index的位置配对
		prevOut := wire.NewOutPoint(prevTxHash, v.FromIndex)
		//组装txin模板
		txIn := wire.NewTxIn(prevOut, nil, nil)
		redeemTx.AddTxIn(txIn)
	}

	//签名
	for i, v := range tpl.TxIns {
		privKey, _ := parsePrivKey(v.FromPrivkey)
		//获取交易脚本
		_, fromPkScript, err := createPayScript(v.FromAddr, coinNetPrams)
		if err != nil {
			return "", fmt.Errorf("get fromAddr,fromPkScript error:%v", err)
		}
		sigScript, err := txscript.SignatureScript(redeemTx, i, fromPkScript, txscript.SigHashAll, privKey, true)
		if err != nil {
			return "", fmt.Errorf("get sigScript error:%v", err)
		}
		redeemTx.TxIn[i].SignatureScript = sigScript

		//校验签名
		flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
			txscript.ScriptStrictMultiSig |
			txscript.ScriptDiscourageUpgradableNops
		vm, err := txscript.NewEngine(fromPkScript, redeemTx, i,
			flags, nil, nil, -1)

		if err != nil {
			return "", fmt.Errorf("check error1:%v", err)
		}
		if err := vm.Execute(); err != nil {
			return "", fmt.Errorf("check error2:%v", err)
		}
	}
	err := checkSign(tpl, redeemTx)
	if err != nil {
		return "", err
	}
	//输出推送的hex
	buf := new(bytes.Buffer)
	redeemTx.Serialize(buf)
	return hex.EncodeToString(buf.Bytes()), nil
}

//创建地址交易脚本
func createPayScript(addrStr string, coinNet *chaincfg.Params) (btcutil.Address, []byte, error) {
	addr, err := btcutil.DecodeAddress(addrStr, coinNet)
	if err != nil {
		return nil, nil, err
	}
	log.Println(addr.EncodeAddress())
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}

//私钥转换，获取返回公私钥
func parsePrivKey(privkeyStr string) (*btcec.PrivateKey, *btcec.PublicKey) {
	wif, _ := btcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

func checkSign(tpl *DogeTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		//redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
		pk, err := txscript.ParsePkScript(v.PkScript)
		if err != nil {
			return err
		} else {
			addr, err := pk.Address(coinNetPrams)
			if err != nil {
				return err
			}
			if addr.EncodeAddress() != tpl.TxOuts[i].ToAddr {
				return fmt.Errorf("签名后数据不一致，before：%s，after：%s", tpl.TxOuts[i].ToAddr, addr.EncodeAddress())
			}
		}

		//_, tplPkScriptByte, err := createPayScript(tpl.TxOuts[i].ToAddr, coinNetPrams)
		//if err != nil {
		//	return fmt.Errorf("finally check vout error：%s", err.Error())
		//}
		//tplPkScript := hex.EncodeToString(tplPkScriptByte)
		//if redeemTxOutAddrPkScript != tplPkScript {
		//	return fmt.Errorf("index:%d,finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", i, redeemTxOutAddrPkScript, tplPkScript)
		//}
		tplam := decimal.NewFromInt(tpl.TxOuts[i].ToAmountInt64)
		if v.Value != tplam.IntPart() {
			return fmt.Errorf("index:%d,finally check vout pkScript error：over amount:%d,before:%d", i, v.Value, tpl.TxOuts[i].ToAmountInt64)
		}
	}
	return nil
}
