package atom

import "testing"

func Test_rpc(t *testing.T) {
	url := "http://atom.rylink.io:20080"
	rpc := NewNodeClient(url)
	t.Log(rpc.AuthAccount("cosmos1un6xp7zkjdcv6ndzecgr9kps8a7rlj9k40thus"))
}
