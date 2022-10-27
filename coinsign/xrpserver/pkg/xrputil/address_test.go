package xrputil

import (
	"testing"
)

func TestGenAddress(t *testing.T) {
	pri, pub, addr, err := GenAddress()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("pri:%s \npub:%s\naddr:%s", pri, pub, addr)
}

func TestGenAddressFromPriv(t *testing.T) {
	//pri:c91f2a7550cacaaea47b85a6ceaa58df952144fde1f19186546fb35afd64a065
	//pub:03364597c28684cc7fb82e70cbc57be96a6e2696af87708f32f6a2739f7ff39c83
	//addr:rarnhJ7JLg6BoCq4TDHZa4mvL614Mgi1Bx
	addr, err := GenAddressFromPriv("c91f2a7550cacaaea47b85a6ceaa58df952144fde1f19186546fb35afd64a065")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("addr:%s", addr)
	t.Log(addr == "rarnhJ7JLg6BoCq4TDHZa4mvL614Mgi1Bx")
}

func TestGenAddressFromSecret(t *testing.T) {
	//sn2a3FtgrqYzHobroyoszXGgBN4Tc
	//rM43k4sCCR9ipy8igUmjtfZk2wNota4y4F
	pri, pub, addr, err := GenAddressFromSecret("sn2a3FtgrqYzHobroyoszXGgBN4Tc")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("pri:%s \npub:%s\naddr:%s", pri, pub, addr)
}
