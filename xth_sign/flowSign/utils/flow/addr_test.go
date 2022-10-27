package flow

import (
	"encoding/hex"
	//"github.com/onflow/flow-go-sdk"
	"testing"
	"encoding/base64"
	//flow "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Test_addr(t *testing.T){
	priBytes,_:=hex.DecodeString("2b803c030c53e0367dd84ed43c8c056351eb3a6d76d656d48f48c33bd70a79a6632a68959d0b31d861e688d8611f960f1285d84449dd351223ee1af78d506ff7")
	priKey,err :=crypto.DecodePrivateKey(crypto.ECDSA_secp256k1,priBytes)
	if err != nil{
		panic(err.Error())
	}
	pubBytes := priKey.PublicKey().Encode()
	//t.Log(hex.EncodeToString(priKey.PublicKey().Encode()))
	base64.StdEncoding.EncodeToString(pubBytes)
}

func Test_base(t *testing.T){
	pubytes,_:=hex.DecodeString("2b803c030c53e0367dd84ed43c8c056351eb3a6d76d656d48f48c33bd70a79a6632a68959d0b31d861e688d8611f960f1285d84449dd351223ee1af78d506ff7")

	t.Log(base64.StdEncoding.EncodeToString(pubytes))
}