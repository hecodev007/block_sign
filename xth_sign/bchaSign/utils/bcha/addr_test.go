package bcha

import (
	"encoding/hex"
	"testing"

	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/txscript"
)

func Test_addr(t *testing.T) {
	//laddr, caddr, pri, err := GenAccount()
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//t.Log(laddr, caddr, pri)
	caddr := "1AnJeoPNk4yzRv5GqoqusbpEtbmLzVJteN"
	addr, err := ToCashAddr(caddr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)

	caddr, err = ToAddr(addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(caddr)
}
func Test_script(t *testing.T) {
	script, err := hex.DecodeString("a914260617ebf668c9102f71ce24aba97fcaaf9c666a87")
	if err != nil {
		t.Fatal(err.Error())
	}
	_, addrs, num, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(num, addrs[0].EncodeAddress())
}
