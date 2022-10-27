package dot

import (
	"encoding/hex"
	sr25519 "github.com/ChainSafe/go-schnorrkel"

	"testing"
)

func Test_acc(t *testing.T){
	pri,pub ,err := GenerateKey()
	t.Log(hex.EncodeToString(pri),pub ,err)
	addr,err :=CreateAddress(pub,PolkadotPrefix)
	t.Log(addr,err)

}
func Test_info(t *testing.T){
	api,err := New("ws://13.231.191.20:9944")
	if err != nil {
		panic(err.Error())
	}
	acc,err :=api.GetAccountInfo("3nbzmMR1aufpfRcmueqatRpBzStubHKxporVw5xz1KNy1KnY")
	if err != nil {
		panic(err.Error())
	}
	t.Log(acc)

}
func Test_pri(t *testing.T){
	prihex :="08dcf472c37978240fefc1f9ce77ad92d0830f95966459d12863ab6e1e23059a"
	//prihex = "2f8a870c797a4d1c28a45d19dd9d3e895a2e1e85fcb3b5e12ecc4945177b4d9a"
	//prihex = "4d2473e4188c3ca5eb878a9a257223af02cb4ff8d409f2529f85f4922ec76c93"
	pribytes ,_ := hex.DecodeString(prihex)
	var tmpribytes [32]byte
	copy(tmpribytes[:],pribytes[:])
	t.Log(hex.EncodeToString(tmpribytes[:]))
	privkey,err := sr25519.NewMiniSecretKeyFromRaw(tmpribytes)
	if err != nil {
		t.Fatal(err.Error())
	}
	pubbytes := privkey.Public().Encode()
	addr,err :=CreateAddress(pubbytes[:],PolkadotPrefix)
	t.Log(addr,err)
}
