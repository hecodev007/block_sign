package btc

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil"
)

func Test_acc(t *testing.T) {
	pri := "L4SdL6gRUfDfDGg2wDptnPxXgacWcRuWq5xmUwnxWu2fjpagyiwg"
	//pri = "KwSVmzdUqcGd5wXW43Ya8geyTGAM9NxyLGnmdjK4m4a5Nm8B2F8i"
	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		panic(err.Error())
	}
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), NetParams)
	if err != nil {
		panic(err.Error())
	}
	address := pkhash.EncodeAddress()
	fmt.Println(address)
	return
}
