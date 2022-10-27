package models

import (
	"encoding/hex"
	"github.com/ystar-foundation/yta-go"
	//"github.com/ystar-foundation/yta-go/ecc"
	"fmt"
	"github.com/ystar-foundation/yta-go/ecc"
	"github.com/ystar-foundation/yta-go/token"
	"testing"
)

func Test_ecc(t *testing.T) {
	wif, err := ecc.NewPrivateKey("5Ki5H9RjiyEJng6uVnD1mdoTJss4jZVWU4yMnA3ZyMVE5AgcKuy")
	if err != nil {
		panic(err.Error())
	}
	t.Log(wif.PublicKey().String())
}

func Test_FixmeTestPackedTransaction_Unpack(t *testing.T) {

	transfer := token.Transfer{}
	fmt.Println(transfer)

	hexString := "4f21a15f0ccc68ccdf9b00000000010000b826630f2ef6000000572d3ccdcd0170558c7a5d9526e500000000a8ed32322570558c7a5d9526e51093ea1ee9a33b61e8030000000000000459535200000000043137373900"
	data, err := hex.DecodeString(hexString)
	if err != nil {
		panic(err.Error())
	}
	tx := yta.PackedTransaction{
		PackedTransaction: data,
	}

	signedTx, err := tx.Unpack()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(signedTx)
}
