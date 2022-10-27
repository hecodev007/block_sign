package alaya

import (
	"encoding/hex"
	"testing"

	"github.com/adiabat/bech32"
)

func Test_acc(t *testing.T) {
	addr, pri, _ := GenAccount()
	t.Log(addr, pri)
	st, bt, _ := bech32.Decode(addr)
	t.Log(st, hex.EncodeToString(bt))
}
func Test_eth(t *testing.T) {
	//priStr := "1beba0161e1b3b187d235df91aca582f1667fcbe57bb97c499d91230d6d4fd10"

}
