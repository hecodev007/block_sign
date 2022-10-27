package rylinkghostutil

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/group-coldwallet/btcserver/model/bo"
	"strings"
)

const (
	VERSION   = int32(160) //版本定制为2
	UNSUPPORT = 0          //暂时不支持,标识
	P2SH      = 1          //定义P2SH地址类型
	P2PKH     = 2          //定义P2PKH地址类型
)

//type TxTpl struct {
//	TxIns []TxInTpl
//	TxOut []TxOutTpl
//}
//
////utxo模板
//type TxInTpl struct {
//	FromAddr         string //来源地址
//	FromPrivkey      string //来源地址地址对于的私钥，签名期间赋值
//	FromTxid         string //来源UTXO的txid
//	FromIndex        uint32 //来源UTXO的txid 地址的下标
//	FromAmount       int64  //来源UTXO的txid 对应的金额
//	FromRedeemScript string //多签脚本
//}
//
////输出模板
//type TxOutTpl struct {
//	ToAddr   string //txout地址
//	ToAmount int64  //txout金额
//}

//构建交易模板,简单校验金额合法性，
//目前根据业务暂时只支持P2PKH（常规1开头）,
func SignTxTpl(tpl *bo.BtcTxTpl) (string, error) {
	if len(tpl.TxIns) < 1 || len(tpl.TxOuts) < 1 {
		return "", errors.New("error input data")
	}
	redeemTx := wire.NewMsgTx(VERSION)
	//组装txout输出
	for _, v := range tpl.TxOuts {
		if !strings.HasPrefix(v.ToAddr, "G") {
			return "", fmt.Errorf("Unsupported  out address type,address:%s", v.ToAddr)
		}
		_, toPkScript, err := createPayScript(v.ToAddr)
		if err != nil {
			return "", err
		}
		//构造txout输出，注意是否存在找零
		txOut := wire.NewTxOut(v.ToAmount, toPkScript)
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
		privKey, _ := ParsePrivKey(v.FromPrivkey)
		//获取交易脚本
		fromAddr, fromPkScript, err := createPayScript(v.FromAddr)
		if err != nil {
			return "", fmt.Errorf("get fromAddr,fromPkScript error:%v", err)
		}
		//判断地址类型，进行各自的签名
		addrType := checkAddressType(fromAddr)
		switch addrType {
		//case P2PKH:
		//pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
		//p2wkhAddr, err := btcutil.NewAddressWitnessPubKeyHash(
		//	pubKeyHash, CoinNet,
		//)
		//if err != nil {
		//	return "", fmt.Errorf("get p2wkhAddr error:%v", err)
		//}
		//witnessProgram, err := txscript.PayToAddrScript(fromAddr)
		//if err != nil {
		//	return "", fmt.Errorf("get witnessProgram error:%v", err)
		//}
		////bldr := txscript.NewScriptBuilder()
		////bldr.AddData(witnessProgram)
		////sigScript, err := bldr.Script()
		////if err != nil {
		////	return "", fmt.Errorf("get sigScript error:%v", err)
		////}
		////redeemTx.TxIn[i].SignatureScript = sigScript
		//hashsign := txscript.NewTxSigHashes(redeemTx)
		//witnessScript, err := txscript.WitnessSignature(redeemTx, hashsign,
		//	i, v.FromAmount, witnessProgram, txscript.SigHashAll, privKey, true,
		//)
		//if err != nil {
		//	return "", fmt.Errorf("get witnessScript error:%v", err)
		//}
		//redeemTx.TxIn[i].Witness = witnessScript

		case P2PKH:
			//常规1地址签名
			//====生成签名方式1 start====
			sigScript, err := txscript.SignatureScript(redeemTx, i, fromPkScript, txscript.SigHashAll, privKey, true)
			//====生成签名方式1 end====

			//====生成签名方式2 start====
			//强制返回注入的私钥
			//lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
			//	return privKey, true, nil
			//}
			//sigScript, err := txscript.SignTxOutput(coinNet,
			//	redeemTx, i, fromPkScript, txscript.SigHashAll,
			//	txscript.KeyClosure(lookupKey), nil, nil)
			//====生成签名方式2 end====

			if err != nil {
				return "", fmt.Errorf("get sigScript error:%v", err)
			}
			redeemTx.TxIn[i].SignatureScript = sigScript
		default:
			//存在暂时不支持的地址类型
			return "", errors.New("There are unsupported address types in it.")

		}
		//校验签名
		//flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		//	txscript.ScriptStrictMultiSig |
		//	txscript.ScriptDiscourageUpgradableNops
		//vm, err := txscript.NewEngine(fromPkScript, redeemTx, i,
		//	flags, nil, nil, -1)
		//
		//if err != nil {
		//	return "", fmt.Errorf("check error1:%v", err)
		//}
		//if err := vm.Execute(); err != nil {
		//	return "", fmt.Errorf("check error2:%v", err)
		//}
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

//私钥转换，获取返回公私钥
func ParsePrivKey(privkeyStr string) (*btcec.PrivateKey, *btcec.PublicKey) {
	wif, _ := btcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

func checkSign(tpl *bo.BtcTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
		if tpl.TxOuts[0].ToUsdtAmount > 0 {

			_, tplPkScriptByte, err := createPayScript(tpl.TxOuts[i-1].ToAddr)
			if err != nil {
				return fmt.Errorf("finally check vout error：%s", err.Error())
			}
			tplPkScript := hex.EncodeToString(tplPkScriptByte)
			if redeemTxOutAddrPkScript != tplPkScript {
				return fmt.Errorf("index:%d,finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", i, redeemTxOutAddrPkScript, tplPkScript)
			}
			if v.Value != tpl.TxOuts[i-1].ToAmount {
				return fmt.Errorf("index:%d,finally check vout pkScript error：over amount:%d,before:%d", i, v.Value, tpl.TxOuts[i].ToAmount)
			}

		} else {
			_, tplPkScriptByte, err := createPayScript(tpl.TxOuts[i].ToAddr)
			if err != nil {
				return fmt.Errorf("finally check vout error：%s", err.Error())
			}
			tplPkScript := hex.EncodeToString(tplPkScriptByte)
			if redeemTxOutAddrPkScript != tplPkScript {
				return fmt.Errorf("index:%d,finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", i, redeemTxOutAddrPkScript, tplPkScript)
			}
			if v.Value != tpl.TxOuts[i].ToAmount {
				return fmt.Errorf("index:%d,finally check vout pkScript error：over amount:%d,before:%d", i, v.Value, tpl.TxOuts[i].ToAmount)
			}
		}

	}
	return nil
}

//创建地址交易脚本
func createPayScript(addrStr string) (btcutil.Address, []byte, error) {
	addr, err := btcutil.DecodeAddress(addrStr, CoinNet)
	if err != nil {
		return nil, nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}

//抽离txscript.PayToAddrScript的方法，判断地址类型
func checkAddressType(addr btcutil.Address) int {
	switch addr := addr.(type) {
	case *btcutil.AddressPubKeyHash:
		if addr == nil {
			return -1
		}
		return P2PKH
	case *btcutil.AddressScriptHash:
		if addr == nil {
			return -1
		}
		return P2SH

	case *btcutil.AddressPubKey:
		if addr == nil {
			return -1
		}
		return UNSUPPORT

	case *btcutil.AddressWitnessPubKeyHash:
		if addr == nil {
			return -1
		}
		return UNSUPPORT
	case *btcutil.AddressWitnessScriptHash:
		if addr == nil {
			return -1
		}
		return UNSUPPORT
	}
	return UNSUPPORT
}
