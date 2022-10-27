package labs

import (
	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/blake2b"

	"encoding/hex"
)

func GenAccount() (addr string, pri string, err error) {
	privatekey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	pri = hex.EncodeToString(privatekey.Serialize())
	pub_bytes := (*btcec.PublicKey)(&privatekey.PublicKey).SerializeCompressed()
	prefix := "secp256k1"
	hash := append([]byte(prefix), 0x0)
	hash = append(hash, pub_bytes...)
	sum256 := blake2b.Sum256(hash)
	addr = hex.EncodeToString(sum256[:])
	return addr, pri, nil
}
