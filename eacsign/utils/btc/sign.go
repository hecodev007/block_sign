package btc

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/gcash/bchutil"
	"strings"
)

func SignTx2(tx *wire.MsgTx, index int, amount int64, pri, FromAddr string) (*wire.MsgTx, error) {
	privKey, pubKey := ParsePrivKey(pri)

	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	p2wkhAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, NetParams)
	if err != nil {
		return nil, fmt.Errorf("get fromAddr,fromPkScript error:%v", err)
	}
	//获取交易脚本
	_, fromPkScript, err := CreatePayScript(FromAddr)

	if err != nil {
		return nil, fmt.Errorf("get p2wkhAddr error:%v", err)
	}
	witnessProgram, err := txscript.PayToAddrScript(p2wkhAddr)
	if err != nil {
		return nil, fmt.Errorf("get witnessProgram error:%v", err)
	}
	bldr := txscript.NewScriptBuilder()
	bldr.AddData(witnessProgram)
	sigScript, err := bldr.Script()
	if err != nil {
		return nil, fmt.Errorf("get sigScript error:%v", err)
	}

	tx.TxIn[index].SignatureScript = sigScript
	hashsign := txscript.NewTxSigHashes(tx)
	witnessScript, err := txscript.WitnessSignature(tx, hashsign,
		index, amount, witnessProgram, txscript.SigHashAll, privKey, true,
	)
	if err != nil {
		return nil, fmt.Errorf("get witnessScript error:%v", err)
	}
	tx.TxIn[index].Witness = witnessScript

	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(fromPkScript, tx, index,
		flags, nil, nil, -1)
	if err != nil {
		return nil, fmt.Errorf("check error1:%v", err)
	}
	if err := vm.Execute(); err != nil {
		return nil, fmt.Errorf("check error2:%v", err)
	}

	//err := checkSign(tpl, redeemTx)
	//if err != nil {
	//	return nil, err
	//}
	//输出推送的hex
	//buf := new(bytes.Buffer)
	//tx.Serialize(buf)
	return tx, nil
}
func SignTx(tx *wire.MsgTx, index int, amount int64, pri string) (*wire.MsgTx, error) {
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		return tx, err
	}
	pk := (*btcec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	pkhash, err := btcutil.NewAddressPubKeyHash(bchutil.Hash160(pk), NetParams)
	if err != nil {
		return tx, err
	}
	from_cashaddr := pkhash.EncodeAddress()

	_, pkScript, err := CreatePayScript(from_cashaddr)
	if err != nil {
		return tx, err
	}
	sigScript, err := txscript.SignatureScript(tx, index, pkScript, txscript.SigHashAll, wif.PrivKey, true)
	if err != nil {
		return tx, err
	}
	tx.TxIn[index].SignatureScript = sigScript
	return tx, nil
}

func SignTx4(tx *wire.MsgTx, index int, amount int64, pri, from_cashaddr string) (*wire.MsgTx, error) {
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		return tx, err
	}
	//pk := (*btcec.PublicKey)(&wif.PrivKey.PublicKey).SerializeCompressed()
	//pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	//if err != nil {
	//	return tx, err
	//}
	//from_cashaddr := pkhash.EncodeAddress()
	addr, pkScript, err := CreatePayScript(from_cashaddr)
	if err != nil {
		return tx, err
	}

	switch addr.(type) {
	case *btcutil.AddressPubKeyHash:
		sigScript, err := txscript.SignatureScript(tx, index, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return tx, err
		}
		tx.TxIn[index].SignatureScript = sigScript
	case *btcutil.AddressWitnessPubKeyHash:
		pubKeyHash := btcutil.Hash160(wif.SerializePubKey())
		p2wkhAddr, err := btcutil.NewAddressWitnessPubKeyHash(
			pubKeyHash, NetParams,
		)
		if err != nil {
			return nil, fmt.Errorf("get p2wkhAddr error:%v", err)
		}
		witnessProgram, err := txscript.PayToAddrScript(p2wkhAddr)
		if err != nil {
			return nil, fmt.Errorf("get witnessProgram error:%v", err)
		}
		bldr := txscript.NewScriptBuilder()
		bldr.AddData(witnessProgram)
		sigScript, err := bldr.Script()
		if err != nil {
			return nil, fmt.Errorf("get sigScript error:%v", err)
		}
		tx.TxIn[index].SignatureScript = sigScript
		hashsign := txscript.NewTxSigHashes(tx)
		witnessScript, err := txscript.WitnessSignature(tx, hashsign,
			index, amount, witnessProgram, txscript.SigHashAll, wif.PrivKey, true,
		)
		if err != nil {
			return nil, fmt.Errorf("get witnessScript error:%v", err)
		}
		tx.TxIn[index].Witness = witnessScript
	default:
		return nil, errors.New("This address is not supported:" + from_cashaddr)
	}

	return tx, nil
}

func SignTx3(tx *wire.MsgTx, index int, FromAddr string, FromAmount int64, pri string) (*wire.MsgTx, error) {
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		return nil, err
	}
	fromAddr, fromPkScript, err := CreatePayScript(FromAddr)
	if err != nil {
		return nil, err
	}
	switch fromAddr.(type) {
	case *btcutil.AddressPubKeyHash:
		sigScript, err := txscript.SignatureScript(tx, index, fromPkScript, txscript.SigHashAll, wif.PrivKey, true)

		if err != nil {
			return nil, fmt.Errorf("get sigScript error:%v", err)
		}
		tx.TxIn[index].SignatureScript = sigScript
	case *btcutil.AddressScriptHash:
		pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
		p2wkhAddr, err := btcutil.NewAddressWitnessPubKeyHash(
			pubKeyHash, NetParams,
		)
		if err != nil {
			return nil, fmt.Errorf("get p2wkhAddr error:%v", err)
		}
		witnessProgram, err := txscript.PayToAddrScript(p2wkhAddr)
		if err != nil {
			return nil, fmt.Errorf("get witnessProgram error:%v", err)
		}
		bldr := txscript.NewScriptBuilder()
		bldr.AddData(witnessProgram)
		sigScript, err := bldr.Script()
		if err != nil {
			return nil, fmt.Errorf("get sigScript error:%v", err)
		}
		tx.TxIn[index].SignatureScript = sigScript
		hashsign := txscript.NewTxSigHashes(tx)
		witnessScript, err := txscript.WitnessSignature(tx, hashsign,
			index, FromAmount, witnessProgram, txscript.SigHashAll, wif.PrivKey, true,
		)
		if err != nil {
			return nil, fmt.Errorf("get witnessScript error:%v", err)
		}
		tx.TxIn[index].Witness = witnessScript
	default:
		return nil, errors.New("不支持的账户类型")
	}

	return tx, nil
}

//创建地址交易脚本
func CreatePayScript(addrStr string) (btcutil.Address, []byte, error) {
	//if err := checkTplAddr(addrStr); err != nil {
	//	return nil, nil, err
	//}

	addr, err := btcutil.DecodeAddress(addrStr, NetParams)
	if err != nil {
		return nil, nil, fmt.Errorf("DecodeAddress error:%s", err.Error())
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, err
	}
	return addr, pkScript, nil
}

func checkTplAddr(addr string) error {
	//if strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") ||
	//	strings.HasPrefix(addr, "q") || strings.HasPrefix(addr, "p") {
	//	return nil
	//}
	if strings.HasPrefix(addr, "S") {
		return nil
	}
	return fmt.Errorf("Unsupported  out address type,address:%s", addr)
}
