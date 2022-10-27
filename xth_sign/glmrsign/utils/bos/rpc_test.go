package bos

import "testing"

func Test_rpc(t *testing.T){
	cli :=  NewRpcClient("https://exchainrpc.okex.org","","")
	balance,err := cli.BalanceOf("0xab0d1578216a545532882e420a8c61ea07b00b12","0x4b3439c8848ee4f42588e16dC679818740194C4a")
	t.Logf(balance.String(),err)
}