package dot

import "testing"

func Test_acc(t *testing.T){
	pri,pub ,err := GenerateKey()
	t.Log(pri,pub ,err)
	addr,err :=CreateAddress(pub,PolkadotPrefix)
	t.Log(addr,err)

}