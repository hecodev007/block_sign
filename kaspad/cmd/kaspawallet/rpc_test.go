package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"testing"
)

func TestRpc(t *testing.T) {
	//10.0.230.18
	clt, err := rpcclient.NewRPCClient("localhost:16210")
	if err != nil {
		fmt.Println(err)
	}
	//resp, err := clt.GetInfo()
	//if err != nil {
	//	fmt.Println(err)
	//}
	resp, err := clt.GetBlockCount()
	//resp, err := clt.GetBalanceByAddress("qqjfmur7cjy77g79ylv02nzffvwmmrctm8kvpgrvku6ctr2avz3jsyrzelv6q")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", resp)
}
