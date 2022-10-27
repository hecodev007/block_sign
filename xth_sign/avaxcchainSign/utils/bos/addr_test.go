package bos

import (
	"github.com/eoscanada/eos-go/ecc"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func Test_addr(t *testing.T){
	pri,err :=ecc.NewRandomPrivateKey()
	if err != nil {
		panic(err.Error())
	}

	t.Log(pri.String())
	pub := pri.PublicKey()

	t.Log(pub.String())
	prikey,err := ethcrypto.HexToECDSA("BFE6626E9B5B6D708944E60BF8A7E517B0CBCFF065402C7D836F921FBD4F43AF")
	if err != nil {
		t.Fatal(err.Error())
	}
	addr := ethcrypto.PubkeyToAddress(prikey.PublicKey)
	t.Log(addr.String())
}

