package dot

import (
	"testing"
	"time"

	"github.com/JFJun/substrate-go/rpc"
)

func Test_rpc(t *testing.T) {
	cli, err := rpc.New("http://18.179.223.150:31833", "", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	h, err := cli.GetFinalizedHead()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(h))
}
func Test_cli(t *testing.T) {
	//node ="http://13.114.44.225:30943"
	//url ="http://13.114.44.225:30960"
	cli := NewRpcClient("http://13.114.44.225:30960", "http://13.114.44.225:30943", "")
	t.Log(cli.PartialFee("0x3d028400787c794c8a29b900fd0ab5465a83fd5af2a8cc08a76baf7a185d5abfe6ba3d4901a4b63659a21598750581c09357cae4291d3b5ab706f9320442ece3ea2ea1595759a07fbae175b4a748c9668454c413524fdfccfacd296b46df7693e3eaf3e58d0000000a030068608c05b08056223df5695f10d431cec8c38f3d860638a9272741182dc1fd670700e8764817", "0x5ffb1a51f6d5f3ef7ec3987e668f5f747de6ddebffb1e9f2f598753de7f77f1c"))

}

func Test_tx(t *testing.T) {
	t.Log(time.Now().Unix())
	cli := NewRpcClient("http://13.230.244.98:31880", "http://18.179.223.150:31833", "")
	//t.Log(cli.GetExtrinsicsByNum(4818100))
	_, err := cli.GetBlock(4817708)
	if err != nil {
		t.Log(err.Error())
	}
}
