package neputil

import (
	"github.com/group-coldwallet/nep5server/util"
	"github.com/shopspring/decimal"
	"math/rand"
	"testing"
)

func TestCreateAddr(t *testing.T) {
	for i := 0; i < 3; i++ {
		t.Log(CreateAddr())
		println(rand.Intn(100))
	}
	dd, _ := decimal.NewFromString("1.23456")
	println(dd.Shift(4).IntPart())

}

func TestValidateNEOAddress(t *testing.T) {
	println(ValidateNEOAddress("ANS84LQwUYFXi4HCjzHBcL6qpn8eJzjubF"))
	util.Demo()
}
