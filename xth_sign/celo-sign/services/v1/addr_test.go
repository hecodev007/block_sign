package v1

import (
	"testing"
	"github.com/ethereum/go-ethereum/common"

)

func Test_addr(t *testing.T){
		addr := common.HexToAddress("0x12bAd172b47287a754048f0d294221a499d1690f")
		t.Log(addr.String())
}