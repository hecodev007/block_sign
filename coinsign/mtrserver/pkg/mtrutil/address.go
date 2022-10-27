package mtrutil

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Address       common.Address
	PrivateKey    *ecdsa.PrivateKey
	PrivateKeyStr string
}

// GenerateAccount generates a random Account
func GenerateAccount() (act *Account, err error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("key generation: ecdsa.GenerateKey failed: " + err.Error())
	}

	return &Account{
		Address:       crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey:    privateKeyECDSA,
		PrivateKeyStr: common.BytesToHash(privateKeyECDSA.D.Bytes()).String(),
	}, nil
}

func GetAccountFromBytes(data []byte) (act *Account, err error) {
	//log.Printf("private : %s ,len : %d", hex.EncodeToString(data), len(data))
	privateKeyECDSA, err := crypto.ToECDSA(data)
	if err != nil {
		return nil, fmt.Errorf("key generation: ecdsa.GenerateKey failed: " + err.Error())
	}

	return &Account{
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}, nil
}

func Base64Encode(data []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return dst
}
