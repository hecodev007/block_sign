package stx

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/common/math"
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
func Test_add(t *testing.T){
	h ,_:=hex.DecodeString("ffffffffffffffffffffffffffffffffffffffff")
	t.Log(	C32_check_encode(1,h))
}
func Test_gen(t *testing.T){
	addr,pri,err := GenAccount2()
	if err != nil{
		panic(err.Error())
	}
	t.Log(addr,pri)
	//    addr_test.go:21: SP1MRQDFDNH3AM6VFYCAHFEJ01G68YFWR4MNRA26C 9a05741dbfeecbac7055aadceb8e192410c15b45b8379d5b503d29be294d97f1
	//addr_test.go:27: SP1Z03AMJRDJT1R9DP0X8GJ2ZXNQQG06RYEAGC2EH KyBWnXv2Y7vrBB8QDtq9VgueVRPk8ciTEf23foMT1tPuM2hkvRbG
}

func Test_btcpri(t *testing.T){
	pri58:="L22mRYkKF1DA82BXjjwwxz2FV6LytS8re5n5TcYFETxmMQGu3odR"
	//SP2MWFT45KQSWQYJ93Q90GAF44GX4D3861VE84PKW
	//SP2MWFT45KQSWQYJ93Q90GAF44GX4D3861VE84PKW
	wif,err :=btcutil.DecodeWIF(pri58)
	if err != nil{
		panic(err.Error())
	}
	pribytes := wif.PrivKey.Serialize()
	pribytes2 :=math.PaddedBigBytes(wif.PrivKey.D, wif.PrivKey.Params().BitSize/8)
	t.Log(hex.EncodeToString(pribytes))
	t.Log(hex.EncodeToString(pribytes2))


	//pri,_ :=ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)

	//t.Log(Json(wif.PrivKey))
	//t.Log(Json(pri))
	pubbytes := wif.SerializePubKey()
	//t.Log(hex.EncodeToString(pribytes))

	t.Log(PriToAddr(hex.EncodeToString(pribytes)))
	t.Log(PubToAddr(pubbytes))
	t.Log(wif.CompressPubKey)
}

func Json(d interface{})string{
	str,_:=json.Marshal(d)
	return string(str)
}