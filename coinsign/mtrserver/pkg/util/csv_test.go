package util

import (
	"github.com/group-coldwallet/mtrserver/pkg/mtrutil"
	"testing"
)

func TestCreateAddrCsv(t *testing.T) {

	var arr []AddrInfo = []AddrInfo{
		AddrInfo{
			Address: "12345a",
			PrivKey: "abcdef",
		},
		AddrInfo{
			Address: "45678b",
			PrivKey: "abcdef",
		},
	}

	addrs, err := CreateAddrCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/mtrserver/pem", "testMchId", "testOrderId1", "mtr", arr, 6, 6)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%+v", addrs)

}

func TestCreateAddrCsv2(t *testing.T) {

	arr := make([]AddrInfo, 0)
	for i := 0; i < 10; i++ {
		acc, _ := mtrutil.GenerateAccount()
		arr = append(arr, AddrInfo{
			Address: acc.Address.String(),
			PrivKey: acc.PrivateKeyStr,
		})
	}

	addrs, err := CreateAddrCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/mtrserver/pem", "testgo", "tesgo123", "mtr", arr, 42, 66)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%+v", addrs)

}
