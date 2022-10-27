package bos

import (
	"github.com/eoscanada/eos-go/ecc"
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
}
