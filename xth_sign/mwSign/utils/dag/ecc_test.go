package dag

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/utils"
	"github.com/ethereum/go-ethereum/crypto"
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
	cfxaddr := string(utils.ToCfxGeneralAddress(addr))
	t.Log(addr.String(),cfxaddr,pri)
}

func Test_imp(t *testing.T){
	ac := sdk.NewAccountManager("keytest")
	addr,err :=ac.ImportKey("cab08bf84e33cdeaca211b3d04e5f3b2d610a713b5a8ee92874edaccb01864c8","123456")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)


}