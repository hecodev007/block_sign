package stx

import (
	"encoding/hex"
	"testing"
)
func Test_FromString(t *testing.T){
	//SP1A2K3ENNA6QQ7G8DVJXM24T6QMBDVS7D0TRTAR5
	//SP1A2K3ENNA6QQ7G8DVJXM24T6QMBDVS7D0TRTAR5
	v,h,err :=C32_check_decode("SP1A2K3ENNA6QQ7G8DVJXM24T6QMBDVS7D0TRTAR5")
	t.Log(PriToAddr("6d430bb91222408e7706c9001cfaeb91b08c2be6d5ac95779ab52c6b431950e001"))
	t.Log(v,hex.EncodeToString(h),err)
	t.Log(	C32_check_encode(v,h))
}

func Test_gen(t *testing.T){
	addr,pri,err := GentAccount()
	if err != nil{
		panic(err.Error())
	}
	t.Log(addr,pri)
	//    addr_test.go:21: SP1MRQDFDNH3AM6VFYCAHFEJ01G68YFWR4MNRA26C 9a05741dbfeecbac7055aadceb8e192410c15b45b8379d5b503d29be294d97f1
}