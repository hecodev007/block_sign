package util

import (
	"crypto/rand"
	"math/big"
)

//两个数字中间的随机数，包含左边，不包含右边
func RandInt64(min, max int64) int64 {
	maxBigInt := big.NewInt(max)
	i, _ := rand.Int(rand.Reader, maxBigInt)
	if i.Int64() < min {
		RandInt64(min, max)
	}
	return i.Int64()
}
