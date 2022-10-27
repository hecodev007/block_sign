package sov

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/ethereum/go-ethereum/crypto"
)

func GenAccount()(addr string,pri string,err error){
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)

}