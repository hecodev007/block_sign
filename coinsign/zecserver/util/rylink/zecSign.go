package rylink

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
	"github.com/group-coldwallet/zecserver/model/bo"
	"github.com/iqoption/zecutil"
	"strings"
)

//区块链网络
var netParams = &chaincfg.MainNetParams

//版本号
var version = int32(4)

//构建签名模板，简单校验金额合法性
//目前只支持普通地址P2PKH
func SignTxTpl(tpl *bo.ZecTxTpl) (string, error) {

	netParams = &chaincfg.MainNetParams
	//map[string][]byte{
	//	scriptAddr.EncodeAddress(): pkScript,
	//}
	if len(tpl.TxIns) == 0 || len(tpl.TxOuts) == 0 {
		return "", errors.New("error input data")
	}
	//if tpl.ExpiryHeight < 500000 {
	//	return "", errors.New("error input ExpiryHeight")
	//}

	//创建签名模板
	newTx := wire.NewMsgTx(version)

	//组装输入
	for _, v := range tpl.TxIns {
		ph, err := chainhash.NewHashFromStr(v.FromTxid)
		if err != nil {
			return "", err
		}
		txIn := wire.NewTxIn(wire.NewOutPoint(ph, uint32(v.FromIndex)), nil, nil)
		newTx.AddTxIn(txIn)

	}
	//组装txout输出
	for _, v := range tpl.TxOuts {
		if !strings.HasPrefix(v.ToAddr, "t1") {
			if !strings.HasPrefix(v.ToAddr, "t3") {
				return "", fmt.Errorf("Unsupported address type,address:%s", v.ToAddr)
				//if !strings.HasPrefix(v.ToAddr, "zs1") {
				//
				//}
			}
		}
		addr, err := zecutil.DecodeAddress(v.ToAddr, "mainnet")
		if err != nil {
			return "", err
		}
		receiverPkScript, err := zecutil.PayToAddrScript(addr)
		txOut := wire.NewTxOut(v.ToAmount, receiverPkScript)
		newTx.AddTxOut(txOut)
	}

	zecTx := &zecutil.MsgTx{
		MsgTx:        newTx,
		ExpiryHeight: uint32(tpl.ExpiryHeight),
	}

	//签名输入
	for i, v := range tpl.TxIns {
		if !strings.HasPrefix(v.FromAddr, "t1") {
			if !strings.HasPrefix(v.FromAddr, "t3") {
				return "", fmt.Errorf("Unsupported address type,address:%s", v.FromAddr)
			}
		}
		wif, err := btcutil.DecodeWIF(v.FromPrivkey)
		if err != nil {
			return "", err
		}
		fromAddr, prevTxScript, err := GetPayScriptByAddr(v.FromAddr)
		if err != nil {
			return "", err
		}
		addrType := CheckAddressType(fromAddr)
		switch addrType {
		case P2SH:
			//暂时不需要支持t3地址签名
			return "", fmt.Errorf("Unsupported address type,address:%s", v.FromAddr)
		case P2PKH:
			sigScript, err := zecutil.SignTxOutput(
				netParams,
				zecTx,
				i,
				prevTxScript,
				txscript.SigHashAll,
				txscript.KeyClosure(func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
					return wif.PrivKey, wif.CompressPubKey, nil
				}),
				nil,
				nil,
				v.FromAmount)
			if err != nil {
				return "", err
			}
			newTx.TxIn[i].SignatureScript = sigScript
		default:
			return "", fmt.Errorf("Unsupported address type,address:%s", v.FromAddr)

		}
	}

	//最终校验地址
	for i, v := range zecTx.MsgTx.TxOut {
		overSignPk := hex.EncodeToString(v.PkScript)
		_, beforSignPk, err := GetPayScriptByAddr(tpl.TxOuts[i].ToAddr)
		if err != nil {
			return "", fmt.Errorf("finally check vout error：%s", err.Error())
		}
		beforPk := hex.EncodeToString(beforSignPk)
		if overSignPk != beforPk {
			return "", fmt.Errorf("finally check vout pkScript error：overPkScript:%s,befor:%s", overSignPk, beforPk)
		}
	}
	err := checkAddr(tpl, zecTx.MsgTx)
	if err != nil {
		return "", nil
	}
	var buf bytes.Buffer
	err = zecTx.ZecEncode(&buf, 0, wire.BaseEncoding)
	if err != nil {
		return "", err
	}
	hexStr := fmt.Sprintf("%x", buf.Bytes())
	if hexStr == "" {
		if err != nil {
			return "", errors.New("sign error")
		}
	}
	return hexStr, nil

}

//创建地址交易脚本
func GetPayScriptByAddr(addrStr string) (btcutil.Address, []byte, error) {
	addr, err := zecutil.DecodeAddress(addrStr, netParams.Name)
	if err != nil {
		return nil, nil, err
	}
	pkScript, err := zecutil.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}

const (
	UNSUPPORT = 0 //暂时不支持,标识
	P2SH      = 1 //定义P2SH地址类型
	P2PKH     = 2 //定义P2PKH地址类型
)

//抽离txscript.PayToAddrScript的方法，判断地址类型
func CheckAddressType(addr btcutil.Address) int {
	switch addr := addr.(type) {
	case *zecutil.ZecAddressPubKeyHash:
		//t1开头
		if addr == nil {
			return -1
		}
		return P2PKH
	case *zecutil.ZecAddressScriptHash:
		//t3开头
		if addr == nil {
			return -1
		}
		return P2SH
	default:
		return UNSUPPORT
	}
}

//私钥转换，获取返回公私钥
func ParsePrivKey(privkeyStr string) (*btcec.PrivateKey, *btcec.PublicKey) {
	wif, _ := btcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

//构建签名模板，简单校验金额合法性
//目前只支持普通地址P2PKH
//错误签名测试
func SignTxTplError(tpl *bo.ZecTxTpl) (string, error) {
	if len(tpl.TxIns) == 0 || len(tpl.TxOuts) == 0 {
		return "", errors.New("error input data")
	}
	if tpl.ExpiryHeight < 500000 {
		return "", errors.New("error input ExpiryHeight")
	}

	//创建签名模板
	newTx := wire.NewMsgTx(version)

	//组装输入
	for _, v := range tpl.TxIns {
		ph, err := chainhash.NewHashFromStr(v.FromTxid)
		if err != nil {
			return "", err
		}
		txIn := wire.NewTxIn(wire.NewOutPoint(ph, uint32(v.FromIndex)), nil, nil)
		newTx.AddTxIn(txIn)
	}

	//组装txout输出
	for _, v := range tpl.TxOuts {
		decoded := base58.Decode(v.ToAddr)
		addr, err := btcutil.NewAddressPubKeyHash(decoded[2:len(decoded)-4], netParams)
		if err != nil {
			return "", err
		}
		receiverPkScript, err := zecutil.PayToAddrScript(addr)
		txOut := wire.NewTxOut(v.ToAmount, receiverPkScript)
		newTx.AddTxOut(txOut)
	}

	zecTx := &zecutil.MsgTx{
		MsgTx:        newTx,
		ExpiryHeight: uint32(tpl.ExpiryHeight),
	}

	//签名输入
	for i, v := range tpl.TxIns {
		fmt.Println(v.FromPrivkey)
		wif, err := btcutil.DecodeWIF(v.FromPrivkey)
		if err != nil {
			return "", err
		}
		_, prevTxScript, err := GetPayScriptByAddr(v.FromAddr)
		if err != nil {
			return "", err
		}
		sigScript, err := zecutil.SignTxOutput(
			netParams,
			zecTx,
			i,
			prevTxScript,
			txscript.SigHashAll,
			txscript.KeyClosure(func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
				return wif.PrivKey, wif.CompressPubKey, nil
			}),
			nil,
			nil,
			v.FromAmount)

		if err != nil {
			return "", err
		}
		newTx.TxIn[i].SignatureScript = sigScript
	}
	//再次校验地址，避免传入的地址跟签名期间转换的地址不一致
	err := checkAddr(tpl, zecTx.MsgTx)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = zecTx.ZecEncode(&buf, 0, wire.BaseEncoding)
	if err != nil {
		return "", err
	}
	hexStr := fmt.Sprintf("%x", buf.Bytes())
	if hexStr == "" {
		if err != nil {
			return "", errors.New("sign error")
		}
	}
	return hexStr, nil

}

func checkAddr(tpl *bo.ZecTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
		_, tplPkScriptByte, err := GetPayScriptByAddr(tpl.TxOuts[i].ToAddr)
		if err != nil {
			return fmt.Errorf("finally check vout error：%s", err.Error())
		}
		tplPkScript := hex.EncodeToString(tplPkScriptByte)
		if redeemTxOutAddrPkScript != tplPkScript {
			return fmt.Errorf("finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", redeemTxOutAddrPkScript, tplPkScript)
		}
		if v.Value != tpl.TxOuts[i].ToAmount {
			return fmt.Errorf("finally check vout pkScript error：over amount:%d,befor:%d", v.Value, tpl.TxOuts[i].ToAmount)
		}
	}
	return nil
}
