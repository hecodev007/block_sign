package dhx

import (
	"encoding/json"
	"testing"
)

func Test_acc(t *testing.T){
	pri,pub ,err := GenerateKey()
	t.Log(pri,pub ,err)
	addr,err :=CreateAddress(pub,PolkadotPrefix)
	t.Log(addr,err)

}

func Test_info(t *testing.T){
	cli,err := New("http://13.230.248.48:22933")
	if err != nil {
		t.Fatal(err.Error())
	}
	info,err :=cli.GetAccountInfo("4PjMG7trqGmkKH2E2j9tDEYHMqoK3JfjsZa4dyUj4eTnAquH")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(info))
	//1000509839416743
	//4297157039621122899836929
}

func Test_info2(t *testing.T){
	cli,err := New("http://13.114.44.225:31933")
	if err != nil {
		t.Fatal(err.Error())
	}
	info,err :=cli.GetAccountInfo("5Hn8xoicWvpJv49yjwm1X7Z5UcHxj1Znj817W76q9WtfGko7")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(info))
}

func String(d interface{}) string{
	str,_:= json.Marshal(d)
	return string(str)
}