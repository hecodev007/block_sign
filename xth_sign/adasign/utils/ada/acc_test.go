package ada

import (
	"encoding/hex"
	"testing"

	"github.com/islishude/bip32"

	"github.com/Bitrue-exchange/libada-go"
	//libada "github.com/Bitrue-exchange/libada-go"
)

func Test_acc(t *testing.T) {
	t.Log(GenAccount())
	t.Log(GenAccount())
	t.Log(GenAccount())
}

func Test_tx2(t *testing.T) {
	pri, _ := hex.DecodeString("a815305e16818bf4989126556b6ebf3d484d754b9123590e55989fd675c44357f99ac2a9fcfd6660e5abb9721265269c14c4fc5e778d7784f862785d32ef81c538c377f4a79e984f88266f33e3379ad0810a724768997d9d19eda02421341a57")
	xprv, _ := bip32.NewXPrv(pri)
	address := libada.NewKeyedEnterpriseAddress(xprv.PublicKey(), libada.Mainnet)
	t.Log(hex.EncodeToString(xprv.PublicKey()), address.String())
	tx := libada.NewTx()

	tx.AddInputs(libada.MustInput("441a7526c9c9d4429e93a8e24af75bbd591299e583fdc6fcaff7df9b78ece996", 0))
	fee := uint64(155381 + 4400*3)
	milkassertid := "8a1cfae21368b8bebbbed9800fec304e95cce39a2a57dc35e2e3ebaa4d494c4b"
	tx.AddOutputs(libada.MustOutputWithAssets("addr1q8e95sy56afr8kj7t6czy6qx6ztghe038p40spwa0xft54wc0h7ch26lev50xa9508swl04n7epcvw7p82cvfgv9xmkqhuzq6c", 300000, milkassertid, 1))
	tx.AddOutputs(libada.MustOutput("addr1vya7d3p36x5gzy2mxddpd9p9azaggxzr0jcmfk33pc3wwas7yk7ry", 1044798-fee))
	tx.SetFee(fee)
	//fee := 155381 + 43946
	//fee := 155381 + 4400*(len(tx.Body.Outputs)+len(tx.Body.Inputs))
	t.Log(hex.EncodeToString(tx.Bytes()))
	tx.AddKeyWitness(libada.NewKeysWitness(xprv.PublicKey(), xprv.Sign(tx.Hash())))
	t.Log(hex.EncodeToString(tx.Hash()))
	t.Log(hex.EncodeToString(tx.Bytes()))
}

//tx := NewTx()
//tx.AddInputs(MustInput("2558aad25ec6b0e74009f36dc60d7fec6602ce43d603e80c9edde9dd54c78eb4", 0))
//tx.AddOutputs(NewOutput(shellAddress, 9000000)).SetFee(1000000).SetInvalidAfter(832163)
//tx.AddKeyWitness(NewKeysWitness(rootpub, rootprv.Sign(tx.Hash())))
//
//fmt.Println(hex.EncodeToString(tx.Bytes()))
