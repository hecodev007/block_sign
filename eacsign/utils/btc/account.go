package btc

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/bech32"
	"math"
	"satSign/common/validator"
	"strings"
)

var NetParams *chaincfg.Params

func init() {
	//NetParams = new(chaincfg.Params)
	//NetParams.PubKeyHashAddrID = 0x00
	//NetParams.ScriptHashAddrID = 0x05
	//NetParams.PrivateKeyID = 0x80
	//NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	//NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}

	//NetParams = new(chaincfg.Params)
	//NetParams.PubKeyHashAddrID = 0x3f
	//NetParams.ScriptHashAddrID = 0x41
	//NetParams.PrivateKeyID = 0x1e
	//NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	//NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	//NetParams.Bech32HRPSegwit = "sat"

	//NetParams = new(chaincfg.Params)
	//NetParams.Name = "main"
	//NetParams.DefaultPort = "46657"
	//NetParams.Bech32HRPSegwit = "bn"

	//eac
	NetParams = new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x5d
	NetParams.ScriptHashAddrID = 0x21
	NetParams.PrivateKeyID = 0xdd
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xAD, 0xE4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xB2, 0x1E}

	chaincfg.Register(NetParams)

}

func GenAccount2() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	pk := (*btcec.PublicKey)(&pri.PublicKey).SerializeCompressed()
	conv, err := bech32.ConvertBits(btcutil.Hash160(pk), 8, 5, true)
	if err != nil {
		return "", "", err
	}
	versionPlusData := make([]byte, 1+len(conv))
	versionPlusData[0] = 0
	for i, d := range conv {
		versionPlusData[i+1] = d
	}
	address, err = bech32.Encode("sat", versionPlusData)
	if err != nil {
		return "", "", err
	}

	return address, wif.String(), nil
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

//TKkuS81difdfDBKq6MBnTSTQ4B5r75PSWp,6479d5b583dc0e695c809e18124ca893c6efa7d59ce1b43feb61acd35806000c
func GenAccountt() {
	pri, _ := btcec.PrivKeyFromBytes(btcec.S256(), []byte("6479d5b583dc0e695c809e18124ca893c6efa7d59ce1b43feb61acd35806000c"))
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	pk := (*btcec.PublicKey)(&pri.PublicKey).SerializeCompressed()
	conv, err := bech32.ConvertBits(btcutil.Hash160(pk), 8, 5, true)
	if err != nil {

	}
	versionPlusData := make([]byte, 1+len(conv))
	versionPlusData[0] = 0
	for i, d := range conv {
		versionPlusData[i+1] = d
	}
	address, err := bech32.Encode("sat", versionPlusData)
	if err != nil {

	}
	fmt.Println(address)
	fmt.Println(wif.String())

}

//func GenAccountUenc() (address string, private string, err error) {
//	privKey, err := btcec.NewPrivateKey(btcec.S256())
//	if err != nil {
//		return "", "", err
//	}
//	p := &chaincfg.MainNetParams
//	p.Name = "main"
//	p.Bech32HRPSegwit = "bn"
//
//	privKeyWif, err := btcutil.NewWIF(privKey, p, true)
//	if err != nil {
//		return "", "", err
//	}
//	pubKeySerial := privKey.PubKey().SerializeCompressed()
//	pubKeyAddress, err := btcutil.NewAddressPubKey(pubKeySerial, p)
//	if err != nil {
//		return "", "", err
//	}
//	wif, err := btcutil.DecodeWIF(privKeyWif.String())
//	if err != nil {
//		return "", "", err
//	}
//
//	return pubKeyAddress.EncodeAddress(), hex.EncodeToString(wif.PrivKey.D.Bytes()), nil
//}

func GenAccountUenc() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	pk := (*btcec.PublicKey)(&pri.PublicKey).SerializeCompressed()
	conv, err := bech32.ConvertBits(btcutil.Hash160(pk), 8, 5, true)
	if err != nil {
		return "", "", err
	}
	versionPlusData := make([]byte, 1+len(conv))
	versionPlusData[0] = 0
	for i, d := range conv {
		versionPlusData[i+1] = d
	}
	address, err = bech32.Encode("bn", versionPlusData)
	if err != nil {
		return "", "", err
	}
	//addr, _, err := CreatePayScript(address)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//switch addr.(type) {
	//case *btcutil.AddressPubKeyHash:
	//	fmt.Println("AddressPubKeyHash")
	//	break
	//case *btcutil.AddressScriptHash:
	//	fmt.Println("AddressScriptHash")
	//	break
	//case *btcutil.AddressWitnessPubKeyHash://this
	//	fmt.Println("AddressWitnessPubKeyHash")
	//	break
	//	}
	return address, wif.String(), nil
}

func GenAccount3() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := wif.SerializePubKey()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), &chaincfg.MainNetParams)
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
	//if inMount - outMount >0 && inMount - outMount <= 2250000{
	if inMount-outMount > 0 && inMount-outMount <= 10000000 {
		if outMount < 546 {
			return nil, errors.New("dust tx")
		}
		return tx, nil
	} else {
		return nil, errors.New("require inMount>outMount and Max fee is 2250000")
	}

}

func BuildTx4(params *validator.SignParams) (tx *wire.MsgTx, err error) {
	tx = wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.Outs {
		//if _, err := btcutil.DecodeAddress(out.ToAddr, &chaincfg.MainNetParams); err != nil {
		if _, err := btcutil.DecodeAddress(out.ToAddr, NetParams); err != nil {
			return nil, err
		}
		outaddr, err := btcutil.DecodeAddress(out.ToAddr, NetParams)
		//outaddr, err := btcutil.DecodeAddress(out.ToAddr, &chaincfg.MainNetParams)
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
			//if _, err := btcutil.DecodeAddress(in.FromAddr, &chaincfg.MainNetParams); err != nil {
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
	//if inMount - outMount >0 && inMount - outMount <= 2250000{
	if inMount-outMount > 0 && inMount-outMount <= 10000000 {
		if outMount < 546 {
			return nil, errors.New("dust tx")
		}
		return tx, nil
	} else {
		return nil, errors.New("require inMount>outMount and Max fee is 10000000")
	}

}

//func BuildTx3(params *validator.SignParams) (tx *wire.MsgTx, err error) {
//	redeemTx := wire.NewMsgTx(1)
//	for _, v := range params.Outs {
//		_, toPkScript, err := CreatePayScript(v.ToAddr)
//		if err != nil {
//			return nil, err
//		}
//		//构造txout输出，注意是否存在找零
//		txOut := wire.NewTxOut(v.ToAmountInt64, toPkScript)
//		redeemTx.AddTxOut(txOut)
//	}
//	for _, v := range params.Ins {
//		prevTxHash, err := chainhash.NewHashFromStr(v.FromTxid)
//		if err != nil {
//			return nil, fmt.Errorf("get prevTxHash error:%v", err)
//		}
//		//构造txin输入，注意index的位置配对
//		prevOut := wire.NewOutPoint(prevTxHash, v.FromIndex)
//		//组装txin模板
//		txIn := wire.NewTxIn(prevOut, nil, nil)
//		redeemTx.AddTxIn(txIn)
//	}
//
//}

func ParsePrivKey(privkeyStr string) (*btcec.PrivateKey, *btcec.PublicKey) {
	wif, _ := btcutil.DecodeWIF(privkeyStr)
	decoded := base58.Decode(wif.String())
	privKeyBytes := decoded[1 : 1+btcec.PrivKeyBytesLen]
	privKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return privKey, pubkey
}

func BuildTx2(params *validator.SignParams) (tx *wire.MsgTx, err error) {
	tx = wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.Outs {
		if strings.HasPrefix(out.ToAddr, "S") ||
			strings.HasPrefix(out.ToAddr, "T") ||
			strings.HasPrefix(out.ToAddr, "sat1") {
			//var addr string
			//if strings.HasPrefix(out.ToAddr, "S") {
			//	addr = out.ToAddr
			//} else {
			//	if addr, err = decodeAddress(out.ToAddr); err != nil {
			//		return nil, err
			//	}
			//}
			outaddr, err := btcutil.DecodeAddress(out.ToAddr, NetParams)
			if err != nil {
				return nil, err
			}
			pubkeyscript, err := txscript.PayToAddrScript(outaddr)
			if err != nil {
				return nil, err
			}
			if out.ToAmountInt64 > math.MaxInt64 {
				return nil, errors.New("out.ToAmount outsize")
			}
			tx.AddTxOut(wire.NewTxOut(out.ToAmountInt64, pubkeyscript))
			outMount += out.ToAmountInt64
			if outMount < 0 {
				return nil, errors.New("out.ToAmount outsize")
			}
		} else {
			return nil, errors.New("不支持该地址类型")
		}
	}
	for _, in := range params.Ins {
		//var addr string
		//if strings.HasPrefix(in.FromAddr, "S") {
		//	addr = in.FromAddr
		//} else {
		//	if addr, err = decodeAddress(in.FromAddr); err != nil {
		//		return nil, err
		//	}
		//}
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
		if in.FromAmountInt64 > math.MaxInt64 {
			return nil, errors.New("in.FromAmount outsize")
		}
		inMount += in.FromAmountInt64
		if inMount < 0 {
			return nil, errors.New("inMount outsize")
		}
	}
	//if inMount - outMount >0 && inMount - outMount <= 2250000{
	if inMount-outMount > 0 && inMount-outMount <= 10000000 {
		if outMount < 546 {
			return nil, errors.New("dust tx")
		}
		return tx, nil
	} else {
		return nil, errors.New("require inMount>outMount and Max fee is 10000000")
	}

}

func decodeAddress(address string) (string, error) {
	decode, bytes, err := bech32.Decode(address)
	if err != nil {
		return "", err
	}
	if decode != "sat" {
		return "", errors.New("地址必须以sat开始")
	}
	bits, err := bech32.ConvertBits(bytes[1:], 5, 8, true)
	if err != nil {
		return "", err
	}
	pkhash, err := btcutil.NewAddressPubKeyHash(bits, NetParams)
	if err != nil {
		return "", err
	}
	addr := pkhash.EncodeAddress()
	return addr, nil
}
