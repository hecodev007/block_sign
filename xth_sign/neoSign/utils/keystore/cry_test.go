package keystore

import "testing"

func Test_cry(t *testing.T){
	en,err :=AesCryptCfb([]byte("123123"),[]byte("pJsFwd9saWmwdexzo0GDFeMl1vMT47qx "),false)
	if err != nil {
		panic(err.Error())
	}
	t.Log(string(en))
}