package zcash

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/HorizenOfficial/rosetta-zen/zenutil"

	"github.com/btcsuite/btcd/txscript"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/wire"

	"github.com/btcsuite/btcd/btcec"

	"github.com/HorizenOfficial/rosetta-zen/zend/chaincfg"
	txscript2 "github.com/HorizenOfficial/rosetta-zen/zend/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
)

func Test_addr(t *testing.T) {
	GenAccount()
	GenAccount()
	addr, pri, err := GenAccount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr, pri)
}

func Test_wif(t *testing.T) {
	private, _ := hex.DecodeString("e86d14dc703fdfef242fde8f7e472c2f515b1383e0b0af0faa5365330dc391a9")
	pri, _ := btcec.PrivKeyFromBytes(btcec.S256(), private)
	wif, err := btcutil.NewWIF(pri, chaincfgParams, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	addr, err := zecutil.Encode(wif.PrivKey.PubKey().SerializeCompressed(), chaincfgParams)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)
	t.Log(wif.String())
}

func Test_decodeaddr(t *testing.T) {
	addr, err := zecutil.DecodeAddress("znSwuhWSWwCDY6ctgPKja5SpYmmctJrErnY", "main")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr.String(), hex.EncodeToString(addr.ScriptAddress()))
	pkscript, err := zecutil.PayToAddrScript(addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("pkscript", hex.EncodeToString(pkscript))
	//76a914146cae2e6f60c208f0c22927c093ff49042cd4f088ac20ba4ff6045c998f9b15b65aa28e527e3e81ee49e76ffb7ebb78587b000000000003b3cd0cb4
	//76a914146cae2e6f60c208f0c22927c093ff49042cd4f088ac
}

func Test_sign(t *testing.T) {
	txid := "123afb78b43f8f92c8e4c3d607521c5a20d342d1d9174bc4fe0932c80c5f907c"
	vout := uint32(1)
	addr := "znkrAHiJARHKec5zCzKS2XVbc3fPZTUG7N3"
	addr = "znSwuhWSWwCDY6ctgPKja5SpYmmctJrErnY"
	amount := int64(1000000)
	tx := wire.NewMsgTx(1)

	//from只支持t1地址
	//if address, err := zecutil.DecodeAddress(addr, "main"); err != nil {
	//	t.Fatal(err.Error())
	//} else if addrType := CheckAddressType(address); addrType != P2PKH {
	//	t.Fatal("unsuport from address: prefix with t1")
	//}
	//
	prevTxHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		t.Fatal(err.Error())
	}
	//构造txin输入，注意index的位置配对
	prevOut := wire.NewOutPoint(prevTxHash, vout)
	//组装txin模板
	txIn := wire.NewTxIn(prevOut, nil, nil)
	tx.AddTxIn(txIn)
	//组装txout输出
	if addr, err := zecutil.DecodeAddress(addr, "main"); err != nil {
		t.Fatal(err.Error())
	} else if addrType := CheckAddressType(addr); addrType == -1 {
		t.Fatal("unsuport to address: prefix with t1 or t3")
	} else if pkScript, err := zecutil.PayToAddrScript(addr); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log(hex.EncodeToString(pkScript))
		pkScript, _ = hex.DecodeString("76a914d8c8a0ac34750410e3f7f6cf8717c7d31d2332b888ac203fa5db9d5fe8d7487a9981f1dfa5bb87040f276f28545817d0b21405000000000340c60ab4")
		txOut := wire.NewTxOut(amount, pkScript)
		tx.AddTxOut(txOut)
	}

	var buf bytes.Buffer
	if err = tx.Serialize(&buf); err != nil {
		t.Fatal(err.Error())
	}
	//tx.
	t.Log(hex.EncodeToString(buf.Bytes()))
	//t.Log(String(tx))
	pkScript, _ := hex.DecodeString("76a914d8c8a0ac34750410e3f7f6cf8717c7d31d2332b888ac203fa5db9d5fe8d7487a9981f1dfa5bb87040f276f28545817d0b21405000000000340c60ab4")
	//e86d14dc703fdfef242fde8f7e472c2f515b1383e0b0af0faa5365330dc391a9
	wif, err := btcutil.DecodeWIF("18nokGXeCx3CYiJzrFvHyVnEZdqCoqoncsSHsKH5MZbDLdyLLK7b")
	if err != nil {
		t.Fatal(err.Error())
	}
	script, err := txscript.SignatureScript(tx, 0, pkScript, txscript.SigHashAll,
		wif.PrivKey, true)
	tx.TxIn[0].SignatureScript = script
	var buf2 bytes.Buffer
	if err = tx.Serialize(&buf2); err != nil {
		t.Fatal(err.Error())
	}
	t.Log(hex.EncodeToString(buf2.Bytes()))
}

func Test_wifA(t *testing.T) {
	tx := wire.NewMsgTx(1)
	bt, err := hex.DecodeString("01000000017c905f0cc83209fec44b17d9d142d3205a1c5207d6c3e4c8928f3fb478fb3a120100000000ffffffff0140420f00000000003f76a914d8c8a0ac34750410e3f7f6cf8717c7d31d2332b888ac203fa5db9d5fe8d7487a9981f1dfa5bb87040f276f28545817d0b21405000000000340c60ab400000000")
	if err != nil {
		t.Fatal(err.Error())
	}
	buff := bytes.NewBuffer(bt)
	err = tx.Deserialize(buff)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(tx))
}

func Test_pay(t *testing.T) {
	pscript, err := PayToAddrScript("zskyTrNtpeMaRCgLXojGFtG5rAkucsAdcVn", "00000000015a31f0360babee15e3133d136ba12bf26f34181f817cb4de91820a", 841291)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(hex.EncodeToString(pscript))

	blockhash, err := hex.DecodeString("00000000015a31f0360babee15e3133d136ba12bf26f34181f817cb4de91820a")
	if err != nil {
		t.Fatal(err.Error())
	}
	toaddress, err := zenutil.DecodeAddress("zskyTrNtpeMaRCgLXojGFtG5rAkucsAdcVn", &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err.Error())
	}
	pkScript, err := txscript2.PayToAddrReplayOutScript(toaddress, blockhash, 841291)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(hex.EncodeToString(pkScript))
}
