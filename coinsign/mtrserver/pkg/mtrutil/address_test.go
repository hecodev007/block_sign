package mtrutil

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zwjlink/meterutil/meter"
	"testing"
)

func TestGenerateAccount(t *testing.T) {
	ac, err := GenerateAccount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(ac.Address.String())
	t.Log(ac.PrivateKeyStr)
	t.Log(common.BytesToHash(ac.PrivateKey.D.Bytes()).String() == ac.PrivateKeyStr)
	t.Log(hexutil.Encode(ac.PrivateKey.D.Bytes()) == ac.PrivateKeyStr)
	t.Log(len(ac.PrivateKey.D.Bytes()))
	t.Log(len(ac.PrivateKeyStr))
	t.Log(len(ac.Address.String()))

}

func TestGetAccountFromBytes(t *testing.T) {

	//address: '0x87eb5D02a42F752923Bf50154fA5cf628276AcDd',
	//privateKey: '0xbca8684dd18823b2847a5f50bb6b61de4a2a92593479946566f478359ebed26b',

	bb, _ := hexutil.Decode("0xbca8684dd18823b2847a5f50bb6b61de4a2a92593479946566f478359ebed26b")
	act, _ := GetAccountFromBytes(bb)
	//0x78d94ab894fc393193c5e5896c8c8fcfc2d09dc2
	t.Log(act.Address.String() == "0x87eb5D02a42F752923Bf50154fA5cf628276AcDd")

	pv, err := crypto.HexToECDSA("66e4215d79bac6a9419bb2acf6edfde14755aecedf6ce40cff33fda4b0adb975")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(common.BytesToHash(pv.D.Bytes()).String())

	t.Log(len("66e4215d79bac6a9419bb2acf6edfde14755aecedf6ce40cff33fda4b0adb975"))
	t.Log(("0x66e4215d79bac6a9419bb2acf6edfde14755aecedf6ce40cff33fda4b0adb975")[2:])
}

func TestCheckAddr(t *testing.T) {
	ac, err := meter.ParseAddress("0x87eb5D02a42F752923Bf50154fA5cf628276AcD11")
	if err != nil {
		t.Log("err0:", err.Error())
	} else {
		t.Log(ac.String())
	}

	ac1, err1 := meter.ParseAddress("87eb5D02a42F752923Bf50154fA5cf628276AcDd")
	if err1 != nil {
		t.Log("err1:", err.Error())
	} else {
		t.Log(ac1.String())
	}

}
