package xrputil

import (
	"encoding/hex"
	"github.com/go-chain/go-xrp/crypto"
)

//参考文档：https://xrpl.org/cryptographic-keys.html#secp256k1-key-derivation
//序列值目前只能为0
func GenAddress() (privkey, pubkey, address string, err error) {
	var addressHash crypto.Hash
	key, err := crypto.GenEcdsaKey()
	if err != nil {
		return "", "", "", err
	}
	var seq0 uint32
	addressHash, err = crypto.AccountId(key, &seq0)
	privkey = hex.EncodeToString(key.Private(&seq0))
	pubkey = hex.EncodeToString(key.Public(&seq0))
	address = addressHash.String()
	return
}

func GenAddressFromPriv(privkey string) (address string, err error) {
	pri, err := hex.DecodeString(privkey)
	if err != nil {
		return "", err
	}
	key := crypto.LoadECDSKey(pri)
	addressHash, err := crypto.AccountId(key, nil)
	if err != nil {
		return "", err
	}
	address = addressHash.String()
	return
}

func GenAddressFromSecret(secret string) (privkey, pubkey, address string, err error) {
	seed, err := crypto.NewRippleHash(secret)
	if err != nil {
		return "", "", "", err
	}
	key, err := crypto.NewECDSAKey(seed.Payload())
	var sequenceZero uint32
	addressHash, err := crypto.AccountId(key, &sequenceZero)
	if err != nil {
		return "", "", "", err
	}
	privkey = hex.EncodeToString(key.Private(&sequenceZero))
	pubkey = hex.EncodeToString(key.Public(&sequenceZero))
	address = addressHash.String()
	return
}
