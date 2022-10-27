package wit

import (
	"crypto/rand"
	"github.com/btcsuite/golangcrypto/bn256"
	"crypto/ecdsa"
)
func GentAccount(){
	ecdsa.GenerateKey(&bn256.{},rand.Reader)
}