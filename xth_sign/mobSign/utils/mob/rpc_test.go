package mob

import "testing"

func Test_rpc(t *testing.T){
	rpc := NewRpcClient("http://18.182.64.34:29090","","")
	//entropy,err :=rpc.Entropy()
	//if err != nil {
	//	panic(err)
	//}
	//t.Log(entropy)
	entropy := "fee0cccd29e321aa091875c0c14eaf5531f8335ccbf2c9551b9d1c09df89d3fa"
	vpri,spri,err :=rpc.GenPri(entropy)
	t.Log(vpri,spri,err)
	monitorid,err := rpc.AddMonitor(vpri,spri,2)
	t.Log(monitorid,err)
	t.Log(rpc.GetMonitor(monitorid))
	//t.Log(rpc.DelMonitor(monitorid))
	t.Log(rpc.GetBalance(monitorid,0))
	t.Log(rpc.GetAddress(monitorid,0))
}