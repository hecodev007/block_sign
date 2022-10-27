package okt

import (
	"github.com/okex/okexchain-go-sdk/utils"
	"testing"
)

func Test_addr(t *testing.T) {
	addr, pri, err := GentAccount()
	if err != nil {
		panic(err.Error())
	}
	t.Log(addr, pri)
}
//func Test_key(t *testing.T){
//	keystore.Base64Decode()
//}
func Test_pri(t *testing.T){
	info,err :=utils.CreateAccountWithPrivateKey("D3C7AF571F95735CF588C087F94D58C1B115A1A8048FAE33D1012F8AF1545F6F","test","")
	if err != nil {
		panic(err.Error())
	}
	t.Log(info.GetAddress().String())
}