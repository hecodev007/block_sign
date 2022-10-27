package dag

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
	"testing"
)

//0x3D3551ccdB0b19aa11bA6c4ed318Ffa0D02CA685 0x1d3551ccdb0b19aa11ba6c4ed318ffa0d02ca685 cab08bf84e33cdeaca211b3d04e5f3b2d610a713b5a8ee92874edaccb01864c8
func Test_ecc(t *testing.T) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err.Error())
	}
	pri :=hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA))
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	cfxaddr := addr.String()
	t.Log(addr.String(),cfxaddr,pri)
}

func Test_imp(t *testing.T){
	ac := sdk.NewAccountManager("keytest",cfxaddress.NetowrkTypeMainnetID)
	addr,err :=ac.ImportKey("cab08bf84e33cdeaca211b3d04e5f3b2d610a713b5a8ee92874edaccb01864c8","123456")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)
}

func Test_addr(t *testing.T){
	addr :=cfxaddress.MustNewFromHex("0x15d80245dc02f5a89589e1f19c5c718e405b56cd",cfxaddress.NetowrkTypeMainnetID)
	t.Log(strings.ToLower(addr.String()))
	addr1,err := cfxaddress.NewFromBase32(strings.ToLower("CFX:TYPE.USER:AAM7UAWF5UBTNMEZVHU9DHC6SGHEA0403YNFNUKA19"))
	if err != nil {
		panic(err.Error())
	}
	t.Log(addr1.String())
}
//abcdefghjkmnprstuvwxyz0123456789
//0123456789ABCDEFGHIJKLMNOPQRSTUV
//1386b4185a223ef49592233b69291bbe5a80c527
//aak2rra2njvd77ezwjvx04kkds9fzagfe6ku8scz91
//abcdefghjkmnprstuvwxyz0123456789