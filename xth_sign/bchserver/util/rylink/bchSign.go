package rylink

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil"
	"github.com/gcash/bchutil/base58"
	"github.com/group-coldwallet/bchserver/model/bo"
	"strings"
)

const (
	UNSUPPORT = 0 //暂时不支持,标识
	P2SH      = 1 //定义P2SH地址类型
	P2PKH     = 2 //定义P2PKH地址类型

)

//构建交易模板,简单校验金额合法性，
//目前根据业务暂时只支持P2PKH（常规1开头）,P2SH（常规3开头，但是要注意3开头不一定都是隔离见证地址,也有多签地址，类型是MS）两种地址的签名
func BchSignTxTpl(tpl *bo.BchTxTpl) (string, error) {
	if len(tpl.TxIns) < 1 || len(tpl.TxOuts) < 1 {
		return "", errors.New("error input data")
	}
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	//组装txout输出
	for _, v := range tpl.TxOuts {
		err := checkTplAddr(v.ToAddr)
		if err != nil {
			return "", err
		}
		_, toPkScript, err := CreatePayScript(v.ToAddr)
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
		txIn := wire.NewTxIn(prevOut, nil)
		redeemTx.AddTxIn(txIn)
	}
	//签名
	for i, v := range tpl.TxIns {
		privKey, _ := parsePrivKey(v.FromPrivkey)
		//获取交易脚本
		fromAddr, fromPkScript, err := CreatePayScript(v.FromAddr)
		if err != nil {
			return "", fmt.Errorf("get fromAddr,fromPkScript error:%v", err)
		}
		//判断地址类型，进行各自的签名
		addrType := checkAddressType(fromAddr)
		switch addrType {
		case P2SH:
			//存在暂时不支持的地址类型
			return "", errors.New("There are unsupported address types in it.")
		case P2PKH:
			//常规1地址签名
			//====生成签名方式1 start====
			sigScript, err := txscript.SignatureScript(redeemTx, i, v.FromAmount, fromPkScript, txscript.SigHashAll, privKey, true)
			//====生成签名方式1 end====
			if err != nil {
				return "", fmt.Errorf("get sigScript error:%v", err)
			}
			redeemTx.TxIn[i].SignatureScript = sigScript
		default:
			//存在暂时不支持的地址类型
			return "", errors.New("There are unsupported address types in it.")

		}
		//校验签名
		flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
			txscript.ScriptVerifySchnorr |
			txscript.ScriptVerifyBip143SigHash
		vm, err := txscript.NewEngine(fromPkScript, redeemTx, i,
			flags, nil, nil, v.FromAmount)
		if err != nil {
			fmt.Println(err)
			return "", fmt.Errorf("i: %d ,check error1:%v", i, err)
		}
		err = vm.Execute()
		if err != nil {
			fmt.Println(err)
			return "", fmt.Errorf("i: %d ,check error2:%v", i, err)
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

//私钥转换，获取返回公私钥
func parsePrivKey(privkeyStr string) (*bchec.PrivateKey, *bchec.PublicKey) {
	wif, _ := bchutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+bchec.PrivKeyBytesLen]
	privKey, pubkey := bchec.PrivKeyFromBytes(bchec.S256(), privKeyBytes)
	return privKey, pubkey
}

//创建地址交易脚本
func CreatePayScript(addrStr string) (bchutil.Address, []byte, error) {
	addr, err := bchutil.DecodeAddress(addrStr, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, fmt.Errorf("DecodeAddress error:%s", err.Error())
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}

//抽离txscript.PayToAddrScript的方法，判断地址类型
func checkAddressType(addr bchutil.Address) int {
	switch addr := addr.(type) {
	case *bchutil.AddressPubKeyHash, *bchutil.LegacyAddressPubKeyHash:
		if addr == nil {
			return -1
		}
		return P2PKH
	case *bchutil.AddressScriptHash, *bchutil.LegacyAddressScriptHash:
		if addr == nil {
			return -1
		}
		return P2SH

	case *bchutil.AddressPubKey:
		if addr == nil {
			return -1
		}
		return UNSUPPORT
	}
	return UNSUPPORT
}

func TestSign() {
	tpl := &bo.BchTxTpl{
		TxIns: []bo.BchTxInTpl{
			bo.BchTxInTpl{
				FromAddr:    "xxx",
				FromPrivkey: "xx",
				FromTxid:    "xx",
				FromIndex:   uint32(1),
				FromAmount:  int64(9200),
			},
		},
		TxOuts: []bo.BchTxOutTpl{
			bo.BchTxOutTpl{
				ToAddr:   "xx",
				ToAmount: int64(8500),
			},
		},
	}
	fmt.Println(BchSignTxTpl(tpl))
}

//func checkSign(tpl *bo.BchTxTpl, redeemTx *wire.MsgTx) error {
//	for i, v := range redeemTx.TxOut {
//		redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
//		_, tplPkScriptByte, err := CreatePayScript(tpl.TxOuts[i].ToAddr)
//		if err != nil {
//			return fmt.Errorf("finally check vout error：%s", err.Error())
//		}
//		tplPkScript := hex.EncodeToString(tplPkScriptByte)
//		if redeemTxOutAddrPkScript != tplPkScript {
//			return fmt.Errorf("finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", redeemTxOutAddrPkScript, tplPkScript)
//		}
//		if v.Value != tpl.TxOuts[i].ToAmount {
//			return fmt.Errorf("finally check vout pkScript error：over amount:%d,befor:%d", v.Value, tpl.TxOuts[i].ToAmount)
//		}
//	}
//	return nil
//}

func checkSign(tpl *bo.BchTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		pk, err := txscript.ParsePkScript(v.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return err
		}
		if strings.HasPrefix(tpl.TxOuts[i].ToAddr, "p") || strings.HasPrefix(tpl.TxOuts[i].ToAddr, "q") {
			addr, err := pk.Address(&chaincfg.MainNetParams)
			if err != nil {
				return err
			}
			if addr.String() != tpl.TxOuts[i].ToAddr {
				return fmt.Errorf("finally check vout error：tpl address:%s, sign out address:%s", tpl.TxOuts[i].ToAddr, addr.String())
			}
		} else if strings.HasPrefix(tpl.TxOuts[i].ToAddr, "1") || strings.HasPrefix(tpl.TxOuts[i].ToAddr, "3") {
			addr, err := pk.Address(&chaincfg.MainNetParams)
			if err != nil {
				return err
			}
			btcaddr, err := ChangeAddressToBtc(addr.String())
			if err != nil {
				return err
			}
			if btcaddr != tpl.TxOuts[i].ToAddr {
				return fmt.Errorf("finally check vout error：tpl address:%s, sign out address:%s", tpl.TxOuts[i].ToAddr, btcaddr)
			}
		} else {
			return fmt.Errorf("error tpl address:%s", tpl.TxOuts[i].ToAddr)
		}
		if v.Value != tpl.TxOuts[i].ToAmount {
			return fmt.Errorf("finally check vout error：tpl amount:%d,sign:%d", tpl.TxOuts[i].ToAmount, v.Value)
		}
	}
	return nil
}

func checkTplAddr(addr string) error {
	if strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") ||
		strings.HasPrefix(addr, "q") || strings.HasPrefix(addr, "p") {
		return nil
	}
	return fmt.Errorf("Unsupported  out address type,address:%s", addr)
}
