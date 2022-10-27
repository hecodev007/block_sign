package dot


import (
	"encoding/hex"
	"errors"
	sr25519 "github.com/ChainSafe/go-schnorrkel"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/blake2b"
)

var (
	SSPrefix = []byte{0x53, 0x53, 0x35, 0x38, 0x50, 0x52, 0x45}
	PolkadotPrefix    = []byte{0x00}

)
func GenAccount()(addr,pri string,err error){
	privkey,pubkey,err := GenerateKey()
	if err != nil{
		return
	}
	addr,err = CreateAddress(pubkey,PolkadotPrefix)
	if err != nil {
		return
	}
	pri = hex.EncodeToString(privkey)
	return
}
func GenerateKey() (privkey []byte, pubkey []byte, err error) {
	secret, err := sr25519.GenerateMiniSecretKey()
	if err != nil {
		return nil, nil, err
	}
	priv := secret.Encode()
	pub := secret.Public().Encode()
	return priv[:], pub[:], nil
}

func CreateAddress(publicKeyHash, prefix []byte) (string, error) {
	if len(publicKeyHash) != 32 {
		return "", errors.New("public hash length is not equal 32")
	}
	payload := appendBytes(prefix, publicKeyHash)
	input := appendBytes(SSPrefix, payload)
	ck := blake2b.Sum512(input)
	checkum := ck[:2]
	address := base58.Encode(appendBytes(payload, checkum))
	if address == "" {
		return address, errors.New("base58 encode error")
	}
	return address, nil
}

func appendBytes(data1, data2 []byte) []byte {
	if data2 == nil {
		return data1
	}
	return append(data1, data2...)
}