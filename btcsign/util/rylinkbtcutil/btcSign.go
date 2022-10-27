package rylinkbtcutil

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
	"github.com/group-coldwallet/btcsign/model/bo"
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
//目前根据业务暂时只支持P2PKH（常规1开头）,P2SH（常规3开头，但是要注意3开头不一定都是隔离见证地址,也有多签地址，类型是MS）两种地址的签名
func SignTxTpl(tpl *bo.BtcTxTpl) (string, error) {
	if len(tpl.TxIns) < 1 || len(tpl.TxOuts) < 1 {
		return "", errors.New("error input data")
	}

	if tpl.TxIns[0].UsdtAmount != 0 {
		if tpl.TxIns[0].UsdtAmount != tpl.TxOuts[1].ToUsdtAmount {
			return "", fmt.Errorf("usdt交易,出账金额不对等，from:%d,out:%d,精度8", tpl.TxIns[0].UsdtAmount, tpl.TxOuts[1].ToUsdtAmount)
		}

		//usdt下标位置判断
		if tpl.TxOuts[1].ToUsdtAmount > 0 {
			if tpl.TxIns[0].UsdtAmount <= 0 {
				return "", fmt.Errorf("usdt交易,但是缺少输入来源")
			}
		}
		if tpl.TxIns[0].UsdtAmount > 0 {
			if tpl.TxOuts[1].ToUsdtAmount <= 0 {
				return "", fmt.Errorf("usdt交易,但是缺少接收地址")
			}
		}
	}

	redeemTx := wire.NewMsgTx(VERSION)
	//组装txout输出
	for i, v := range tpl.TxOuts {
		if !strings.HasPrefix(v.ToAddr, "1") {
			if !strings.HasPrefix(v.ToAddr, "3") {
				if !strings.HasPrefix(v.ToAddr, "bc") {
					return "", fmt.Errorf("Unsupported  out address type,address:%s", v.ToAddr)
				}
			}
		}
		if v.ToUsdtAmount != 0 && i != 1 {
			return "", fmt.Errorf("输出的USDT地址下标不在1的位置，%s,下标:%d", v.ToAddr, i)
		}
		if i == 1 && v.ToUsdtAmount != 0 {
			omnihex := createUsdtInfo(v.ToUsdtAmount)
			usdthex, _ := hex.DecodeString(omnihex)
			test, _ := txscript.NullDataScript(usdthex)
			output := wire.NewTxOut(0, test)
			redeemTx.AddTxOut(output)
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
	for i, v := range tpl.TxIns {
		if v.UsdtAmount != 0 && i != 0 {
			return "", fmt.Errorf("输入的USDT地址下标不在0的位置，%s,下标：%d", v.FromAddr, i)
		}
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
			pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
			p2wkhAddr, err := btcutil.NewAddressWitnessPubKeyHash(
				pubKeyHash, CoinNet,
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
	wif, _ := btcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

//创建地址交易脚本
func CreatePayScript(addrStr string) (btcutil.Address, []byte, error) {
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
func CheckAddressType(addr btcutil.Address) int {
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

func main() {
	fromAddr, _ := btcutil.DecodeAddress("3QVbWNWRQuN2Xiv93327h7kZQMVacNqB5d", CoinNet)
	fmt.Println(CheckAddressType(fromAddr))
	tpl := &bo.BtcTxTpl{
		TxIns: []bo.BtcTxInTpl{

			//bo.BtcTxInTpl{
			//	FromAddr:    "362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
			//	FromPrivkey: "L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg",
			//	FromTxid:    "23c1a54e5e0f96b4f32542d1611fd9cf4e75d724966025abf4e6f7a00ff8a5bc",
			//	FromIndex:   uint32(1),
			//	FromAmount:  int64(57248),
			//},
			bo.BtcTxInTpl{
				FromAddr:    "1Mysym482ixvWxifnTT3VoAEJrQBC6Gjga",
				FromPrivkey: "Kzs9fN8fpi14rrgBvFNA8oKFeqNtj9LhrncJcfx1m6FUvKPKZRqo",
				FromTxid:    "558748dc59a4fed4e8db55b5c9d6a0019e24f1b636fcc730db1fb9674c6f565b",
				FromIndex:   uint32(0),
				FromAmount:  int64(38212),
			},
			//TxInTpl{
			//	FromAddr:"362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
			//	FromPrivkey:"L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg",
			//	FromTxid:"d7828a4e2668a226122924e101d6d3605fd6d35beb878bf04e87d8c51676864c",
			//	FromIndex:uint32(3),
			//	FromAmount:int64(461248),
			//},
			//TxInTpl{
			//	FromAddr:"1Mysym482ixvWxifnTT3VoAEJrQBC6Gjga",
			//	FromPrivkey:"Kzs9fN8fpi14rrgBvFNA8oKFeqNtj9LhrncJcfx1m6FUvKPKZRqo",
			//	FromTxid:"aabdeffe7608d8ac74d63b88dc1a38e204e25fedcca75f7f061607d2dc94332d",
			//	FromIndex:uint32(0),
			//	FromAmount:int64(24774),
			//},
			//TxInTpl{
			//	FromAddr:"1Mysym482ixvWxifnTT3VoAEJrQBC6Gjga",
			//	FromPrivkey:"Kzs9fN8fpi14rrgBvFNA8oKFeqNtj9LhrncJcfx1m6FUvKPKZRqo",
			//	FromTxid:"f9ae074c43f596136b8ad39e338c73a7580491af597c7183c754f561cd1d11de",
			//	FromIndex:uint32(2),
			//	FromAmount:int64(546),
			//},
		},
		TxOuts: []bo.BtcTxOutTpl{
			//TxOutTpl{
			//	ToAddr:"3FnBLCSBG8cWWgbwFkbrw294wdmKAo8Tin",
			//	ToAmount:int64(5000),
			//},
			//TxOutTpl{
			//	ToAddr:"33WGKr1VS58TgFBi4NGG5LzSKLsMuX5vNV",
			//	ToAmount:int64(5000),
			//},
			//TxOutTpl{
			//	ToAddr:"362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
			//	ToAmount:int64(187472),
			//},
			//TxOutTpl{
			//	ToAddr:"362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
			//	ToAmount:int64(57248),
			//},
			//TxOutTpl{
			//	ToAddr:"19fse5DwT4jBMBM2j99ctruTqdP9MmknAq",
			//	ToAmount:int64(546),
			//},
			//TxOutTpl{
			//	ToAddr:"1Mysym482ixvWxifnTT3VoAEJrQBC6Gjga",
			//	ToAmount:int64(26774),
			//},
			bo.BtcTxOutTpl{
				ToAddr:   "362c5QyjhHKt92xKAg3X8hHJL8USeqqbDU",
				ToAmount: int64(20000),
			},
			bo.BtcTxOutTpl{
				ToAddr:   "1Mysym482ixvWxifnTT3VoAEJrQBC6Gjga",
				ToAmount: int64(13212),
			},
		},
	}

	fmt.Println(SignTxTpl(tpl))
}

func createUsdtInfo(usdtAmont int64) string {
	amount := fmt.Sprintf("%x", usdtAmont)
	//fmt.Println(addPreZero2(amount))
	usdtcoin := "6f6d6e69" + "000000000000001f" + addPreZero2(amount)
	return usdtcoin
}

func addPreZero2(num string) string {
	ln := len(num)
	s := ""
	for i := 0; i < 16-ln; i++ {
		s += "0"
	}
	return s + num
}

func checkSign(tpl *bo.BtcTxTpl, redeemTx *wire.MsgTx) error {
	for i, v := range redeemTx.TxOut {
		redeemTxOutAddrPkScript := hex.EncodeToString(v.PkScript)
		if len(tpl.TxOuts) > 0 {
			if len(tpl.TxOuts) > 2 && tpl.TxOuts[1].ToUsdtAmount > 0 {
				if strings.HasPrefix(redeemTxOutAddrPkScript, "6a146f6d6e69000000000000001f") {
					//usdt验证
					scriptStr := strings.Replace(redeemTxOutAddrPkScript, "6a146f6d6e69000000000000001f", "", -1)
					usdtStr := createUsdtInfo(tpl.TxOuts[1].ToUsdtAmount)
					usdtStr = strings.Replace(usdtStr, "6f6d6e69000000000000001f", "", -1)
					if scriptStr != usdtStr {
						return fmt.Errorf("finally check vout error usdtAmount sign：%s,before:%s", scriptStr, usdtStr)
					}
					continue
				}
			}
			//_, tplPkScriptByte, err := CreatePayScript(tpl.TxOuts[2].ToAddr)
			//if err != nil {
			//	return fmt.Errorf("finally check vout error：%s", err.Error())
			//}
			//tplPkScript := hex.EncodeToString(tplPkScriptByte)
			//if redeemTxOutAddrPkScript != tplPkScript {
			//	return fmt.Errorf("index:%d,finally check vout pkScript error：rTx PkScript:%s,tplPkScript:%s", i, redeemTxOutAddrPkScript, tplPkScript)
			//}
			//if v.Value != tpl.TxOuts[2].ToAmount {
			//	return fmt.Errorf("index:%d,finally check vout pkScript error：over amount:%d,before:%d", i, v.Value, tpl.TxOuts[i].ToAmount)
			//}

		} else {
			_, tplPkScriptByte, err := CreatePayScript(tpl.TxOuts[i].ToAddr)
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
