package keystore

import "testing"

func Test_key(t *testing.T){
	tmpa:=[]byte("okGiqbil32Z4O22qtcee6Vao1ztoGuxTHPV10DiE+XAYbPZ+4eZTiQc0bJgWQjayjWwLOFPteFTg8A/fHjPU3g==")
	bkey:=[]byte("Z9EFSTk26TQxx+Qv9g58gUBaT+Lmh3mT")
	prikey := []byte("00E4A0137EA1F33413787F3DD60DEC6006914BC685F2E430AF811F46B4FCC849")
	pri,err:=AesBase64CryptCfb(prikey,bkey,true)
	if err !=nil{
		panic(err.Error())
	}

	t.Log(string(pri))

	akey,err :=Base64Decode(tmpa)
	if err != nil {
		panic(err.Error())
	}
	pri,_=AesCryptCfb(akey,bkey,false)
	t.Log(string(pri))
}