package rylink

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	btccfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/group-coldwallet/ltcserver/model/bo"
	"github.com/zwjlink/ltcd/btcec"
	"github.com/zwjlink/ltcd/chaincfg"
	"github.com/zwjlink/ltcd/chaincfg/chainhash"
	"github.com/zwjlink/ltcd/txscript"
	"github.com/zwjlink/ltcd/wire"
	"github.com/zwjlink/ltcutil"
	"github.com/zwjlink/ltcutil/base58"
	"strings"
)

const (
	VERSION   = int32(2) //版本定制为2
	UNSUPPORT = 0        //暂时不支持,标识
	P2SH      = 1        //定义P2SH地址类型
	P2PKH     = 2        //定义P2PKH地址类型

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
func LtcSignTxTpl(tpl *bo.LtcTxTpl) (string, error) {

	if len(tpl.TxIns) < 1 || len(tpl.TxOuts) < 1 {
		return "", errors.New("error input data")
	}

	//更改输出的ltc 3开头地址
	for i, v := range tpl.TxOuts {
		if strings.HasPrefix(v.ToAddr, "3") {
			//转换为M地址
			ltcAddr, err := ChangeAddrBtcToLtc(v.ToAddr)
			if err != nil {
				return "", err
			}
			tpl.TxOuts[i].ToAddr = ltcAddr
		}
	}

	redeemTx := wire.NewMsgTx(VERSION)
	//组装txout输出
	for _, v := range tpl.TxOuts {
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
		txIn := wire.NewTxIn(prevOut, nil, nil)
		redeemTx.AddTxIn(txIn)
	}
	//签名
	for i, v := range tpl.TxIns {
		privKey, pubKey := ParsePrivKey(v.FromPrivkey)
		//获取交易脚本
		fromAddr, fromPkScript, err := CreatePayScript(v.FromAddr)
		if err != nil {
			return "", fmt.Errorf("get fromAddr,fromPkScript error:%v", err)
		}
		//判断地址类型，进行各自的签名
		addrType := CheckAddressType(fromAddr)
		switch addrType {
		case P2SH:
			//3开头隔离见证地址签名
			pubKeyHash := ltcutil.Hash160(pubKey.SerializeCompressed())
			p2wkhAddr, err := ltcutil.NewAddressWitnessPubKeyHash(
				pubKeyHash, &chaincfg.MainNetParams,
			)
			if err != nil {
				return "", fmt.Errorf("get p2wkhAddr error:%v", err)
			}
			witnessProgram, err := txscript.PayToAddrScript(p2wkhAddr)
			if err != nil {
				return "", fmt.Errorf("get witnessProgram error:%v", err)
			}
			bldr := txscript.NewScriptBuilder()
			bldr.AddData(witnessProgram)
			sigScript, err := bldr.Script()
			if err != nil {
				return "", fmt.Errorf("get sigScript error:%v", err)
			}
			redeemTx.TxIn[i].SignatureScript = sigScript
			hashsign := txscript.NewTxSigHashes(redeemTx)
			witnessScript, err := txscript.WitnessSignature(redeemTx, hashsign,
				i, v.FromAmount, witnessProgram, txscript.SigHashAll, privKey, true,
			)
			if err != nil {
				return "", fmt.Errorf("get witnessScript error:%v", err)
			}
			redeemTx.TxIn[i].Witness = witnessScript

		case P2PKH:
			//常规1地址签名
			//====生成签名方式1 start====
			sigScript, err := txscript.SignatureScript(redeemTx, i, fromPkScript, txscript.SigHashAll, privKey, true)
			//====生成签名方式1 end====

			//====生成签名方式2 start====
			//强制返回注入的私钥
			//lookupKey := func(a ltcutil.Address) (*btcec.PrivateKey, bool, error) {
			//	return privKey, true, nil
			//}
			//sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams,
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

//私钥转换，获取返回公私钥
func ParsePrivKey(privkeyStr string) (*btcec.PrivateKey, *btcec.PublicKey) {
	wif, _ := ltcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

//创建地址交易脚本
func CreatePayScript(addrStr string) (ltcutil.Address, []byte, error) {

	if strings.HasPrefix(addrStr, "L") || strings.HasPrefix(addrStr, "M") {
		addr, err := ltcutil.DecodeAddress(addrStr, &chaincfg.MainNetParams)
		if err != nil {
			return nil, nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, nil, err
		}
		return addr, pkScript, nil
	} else {
		if strings.HasPrefix(addrStr, "3") {
			fmt.Println(addrStr)
			fmt.Println(addrStr)
			fmt.Println(addrStr)
			addra, _ := btcutil.DecodeAddress(addrStr, &btccfg.MainNetParams)
			fmt.Println(addra.EncodeAddress())
			fmt.Println(addra.String())
			addr, err := ltcutil.DecodeAddress(addrStr, &chaincfg.MainNetParams)
			if err != nil {
				fmt.Println("123123123")
				return nil, nil, err
			}

			pkScript, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return nil, nil, err
			}
			fmt.Println(addrStr)
			return addr, pkScript, nil
		} else {
			return nil, nil, errors.New("error address type:" + addrStr)
		}
	}

}

//抽离txscript.PayToAddrScript的方法，判断地址类型
func CheckAddressType(addr ltcutil.Address) int {
	switch addr := addr.(type) {
	case *ltcutil.AddressPubKeyHash:
		if addr == nil {
			return -1
		}
		return P2PKH
	case *ltcutil.AddressScriptHash:
		if addr == nil {
			return -1
		}
		return P2SH

	case *ltcutil.AddressPubKey:
		if addr == nil {
			return -1
		}
		return UNSUPPORT

	case *ltcutil.AddressWitnessPubKeyHash:
		if addr == nil {
			return -1
		}
		return UNSUPPORT
	case *ltcutil.AddressWitnessScriptHash:
		if addr == nil {
			return -1
		}
		return UNSUPPORT
	}
	return UNSUPPORT
}

func testmain1() {
	fromAddr, _ := ltcutil.DecodeAddress("3QVbWNWRQuN2Xiv93327h7kZQMVacNqB5d", &chaincfg.MainNetParams)
	fmt.Println(CheckAddressType(fromAddr))
	tpl := &bo.LtcTxTpl{
		TxIns: []bo.LtcTxInTpl{
			bo.LtcTxInTpl{
				FromAddr:    "362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
				FromPrivkey: "L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg",
				FromTxid:    "23c1a54e5e0f96b4f32542d1611fd9cf4e75d724966025abf4e6f7a00ff8a5bc",
				FromIndex:   uint32(1),
				FromAmount:  int64(57248),
			},
		},
		TxOuts: []bo.LtcTxOutTpl{
			bo.LtcTxOutTpl{
				ToAddr:   "3QVbWNWRQuN2Xiv93327h7kZQMVacNqB5d",
				ToAmount: int64(53248),
			},
		},
	}

	fmt.Println(LtcSignTxTpl(tpl))
}

func checkSign(tpl *bo.LtcTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
		_, tplPkScriptByte, err := CreatePayScript(tpl.TxOuts[i].ToAddr)
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
