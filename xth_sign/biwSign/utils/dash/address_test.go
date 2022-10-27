package dash

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil"
)

func Test_addr(t *testing.T) {
	addr, pri, err := GenAccount()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(addr, pri)
	address, err := btcutil.DecodeAddress("XeTK75cqUMwHrsN4ao4V8U8BUbxPYa3c3y", NetParams)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf(address.String())

}
