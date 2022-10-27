package kava

import (
	"terrasign/utils/keystore"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func Test_sdk(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	// Build the account
	t.Log(acc.String())
}
func Test_acc(t *testing.T) {
	addr, pri, err := GenAccount()
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(addr, pri)
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
