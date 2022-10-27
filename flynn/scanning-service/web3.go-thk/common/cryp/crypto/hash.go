package crypto

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/common/cryp/sha3"
	"hash"
)

// 可以计算Hash值的接口类型
type Hasher interface {
	HashValue() ([]byte, error)
}

func Hash256s(in ...[]byte) ([]byte, error) {
	return Keccak256(in...), nil
}

func GetHash256() hash.Hash {
	return sha3.NewKeccak256()
}

func Verify(pub []byte, hash []byte, sig []byte) bool {
	if len(pub) != 65 || len(sig) != 65 {
		return false
	}
	return VerifySignature(pub, hash, sig[:64])
}
