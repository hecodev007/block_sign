package rylink

import (
	"fmt"
	"testing"
)

func TestCreateAddress(t *testing.T) {
	fmt.Println(CreateAddress())
}

func TestChangeAddrToBTC(t *testing.T) {
	t.Log(ChangeAddrBtcToLtc("38xM8CWkTGRKoYZoHDHDi881dwT8K9Yrc5"))
	t.Log(ChangeAddrBtcToLtc("37TszUAK97mcWeR3BJWD6uR2miGVfJNVvN"))
	t.Log(ChangeAddrBtcToLtc("MFAVS5viQPGkc3qhP6GZXmNQxe3aJ8w3Pf"))
	t.Log(ChangeAddrLtcToBTC("MFAVS5viQPGkc3qhP6GZXmNQxe3aJ8w3Pf"))
}
