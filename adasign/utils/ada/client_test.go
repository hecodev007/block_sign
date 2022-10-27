package ada

import (
	"encoding/json"
	"testing"

	"github.com/onethefour/common/xutils"
)

func Test_rpc(t *testing.T) {
	cli := NewRpcClient("http://54.250.240.45:8080", "", "")
	block, err := cli.GetBlockCount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(block))
	amount, err := cli.BalanceOf("addr1vxggvx6uq9mtf6e0tyda2mahg84w8azngpvkwr5808ey6qsy2ww7d", "", 0)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(amount.String())
	unspend, err := cli.Coins("addr1vxggvx6uq9mtf6e0tyda2mahg84w8azngpvkwr5808ey6qsy2ww7d", true)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(unspend))
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
