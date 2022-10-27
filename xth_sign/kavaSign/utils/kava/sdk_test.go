package kava

import (
	"encoding/hex"
	"kavaSign/utils/keystore"
	"testing"

	"github.com/onethefour/common/xutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/kava-labs/kava/app"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func Test_sdk(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	// Build the account
	t.Log(acc.String())

}

func Test_sendTx(t *testing.T) {
	//kava1082p54j743425md04hgpc8ft68j2jvkt7xsefz,KHXxY+4sVLYHoA0qIhX+YnQTxLA6AB/+GZlJV7g+1fE=
	pri, err := keystore.Base64Decode([]byte("KHXxY+4sVLYHoA0qIhX+YnQTxLA6AB/+GZlJV7g+1fE="))
	if err != nil {
		t.Fatal(err.Error())
	}
	var privKey secp256k1.PrivKeySecp256k1
	copy(privKey[:], pri[:])
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	t.Log(acc.String())
}

func Test_cli(t *testing.T) {
	cli := NewNodeClient("http://13.114.44.225:30680/")
	//t.Log(cli.AuthBalance("kava1xd39avn2f008jmvua0eupg39zsp2xn3wf802vn"))
	t.Log(cli.AuthAccount("kava1xd39avn2f008jmvua0eupg39zsp2xn3wf802vn"))
}

func Test_tx_decoder(t *testing.T) {
	codec := app.MakeCodec()
	txBytes, _ := hex.DecodeString("d101282816a90a3fa8a3619a0a148884c8fb2d991b8bf600e4829b60fc526341ea321214ad9890d03e2be95b34e207d4136021735e741fb01a0d0a05756b61766112043130303012130a0d0a05756b61766112043235303010c0843d1a6a0a26eb5ae98721032483fe9fd88a0a34bbddb30e2f6ec3aa8978013a3e514b7a4ffa4d49a589979e1240cd891c5ad13f8dde4a6e4db133010dbb4e42110b0421131c7984f13f67adc38c483119337d399e0dde684722d8698cd3fdfe93b368acb29dfc6c52850da4de3f2209746573743132333435")
	tx, err := auth.DefaultTxDecoder(codec)(txBytes)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(len(tx.GetMsgs()))
	t.Log(tx.GetMsgs()[0].Route(), tx.GetMsgs()[0].Type()) //bank send
	t.Log(xutils.String(tx))
}
