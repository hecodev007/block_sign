package sgb

import (
	"cphsign/utils/sha3"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/ed25519"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

func StringToPrivateKey(privateKeyStr string) (ed25519.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	return privateKey, nil
}
func ToCommonAddress(addr string) (address common.Address) {
	addr = strings.Replace(strings.ToLower(addr), "cph", "0x", 1)
	return common.HexToAddress(addr)
}

//
//// SignTx signs the transaction using the given signer and private key
//func SignTx2(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
//	h := s.Hash(tx)
//	sig, err := crypto.Sign(h[:], prv)
//	if err != nil {
//		return nil, err
//	}
//	return tx.WithSignature(s, sig)
//}
