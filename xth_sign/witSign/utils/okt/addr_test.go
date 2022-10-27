package okt

import "testing"

func Test_addr(t *testing.T) {
	addr, pri, err := GentAccount()
	if err != nil {
		panic(err.Error())
	}
	t.Log(addr, pri)
}
