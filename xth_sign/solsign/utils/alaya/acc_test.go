package alaya

import (
	"encoding/hex"
	"testing"
	"github.com/adiabat/bech32"

)

func Test_acc(t *testing.T){
	addr,pri,_:=GenAccount()
	t.Log(addr,pri)
	st,bt,_:=bech32.Decode(addr)
	t.Log(st,hex.EncodeToString(bt))
}