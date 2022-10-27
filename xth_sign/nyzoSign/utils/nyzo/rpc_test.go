package nyzo

import (
	"testing"
)

func Test_rpc(t *testing.T) {
	rpc := NewRpcClient("http://3.114.68.247:4000", "", "")
	//t.Log(rpc.Info())
	from := "id__8fJXg0Xg3.yB3Gqk.oqvR2grRhsBN-HWjrSoTx9tHsargHdzaTjV"
	pri := "key_83yW~fedJamCNMCnXU6_wfs8_aKQ1q7ME6eLN.GjjwE4pGschDpZ"
	to := "id__8eIF~7YYNBej_XmkQ.b4XTGHMqnsev2X98V3LnToX~u.o-xSRKEi"
	memo := "test123456"
	amount := uint64(1)
	broadcast := false
	t.Log(rpc.SendTransaction(from, to, amount, memo, pri, broadcast))
}
