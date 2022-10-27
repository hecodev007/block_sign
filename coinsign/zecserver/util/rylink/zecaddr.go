package rylink

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
)

//生成地址返回地址和私钥
func CreateAddress() (addr, privKey string, e error) {
	var (
		privWif *btcutil.WIF
		err     error
	)
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	privWif, err = btcutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", err
	}
	privkey := privWif.String()
	var encodedAddr string
	encodedAddr, err = zecutil.Encode(privWif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.Params{
		Name: "mainnet",
	})
	if err != nil {
		return "", "", err
	}
	return encodedAddr, privkey, nil
}

//生成地址返回地址和私钥
func CreateAddress2() (addr, segwwAddr, privKey string, e error) {
	var (
		privWif *btcutil.WIF
		err     error
	)
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", "", err
	}
	privWif, err = btcutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", "", err
	}
	privkey := privWif.String()

	addr, err = zecutil.Encode(privWif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", err
	}

	script := pubKeyHashToScript(privWif.PrivKey.PubKey().SerializeCompressed())
	zecutil.Encode(script, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", err
	}

	a1, pk1, _ := GetPayScriptByAddr(addr)
	a2, pk2, _ := GetPayScriptByAddr(segwwAddr)

	println(a1.EncodeAddress())
	println(a2.EncodeAddress())

	_, a11, _, _ := txscript.ExtractPkScriptAddrs(pk1, &chaincfg.MainNetParams)
	_, a22, _, _ := txscript.ExtractPkScriptAddrs(pk2, &chaincfg.MainNetParams)

	println(a11[0].EncodeAddress())
	println(a22[0].EncodeAddress())

	return addr, segwwAddr, privkey, nil
}

func pubKeyHashToScript(pubKey []byte) []byte {
	pubKeyHash := btcutil.Hash160(pubKey)
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).AddData(pubKeyHash).Script()
	if err != nil {
		panic(err)
	}
	return script
}
