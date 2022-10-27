package util

import "testing"

func TestCreateAddrCsv(t *testing.T) {

	var arr []AddrInfo = []AddrInfo{
		AddrInfo{
			Address: "123123123123",
			PrivKey: "abcdabcdabcdabcdabcd",
		},
		AddrInfo{
			Address: "456456456456456456",
			PrivKey: "qwertyuiqwertyui",
		},
	}
	addrs, err := CreateAddrCsv("/Users/tmp", "testMchId", "testOrderId", "uca", arr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%+v", addrs)

}
