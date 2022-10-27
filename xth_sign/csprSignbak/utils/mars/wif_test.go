package mars

import (
	"testing"
)


func Test_wif(t *testing.T){
	t.Log("123")
	address,pri,err := GenAccount()
	if err != nil{
		panic(err.Error())
	}
	t.Log(address,pri)
}