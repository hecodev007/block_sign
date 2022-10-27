

package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type ComplexIntParameter int64

func (s ComplexIntParameter) ToHex() string {

	return fmt.Sprintf("0x%x", s)

}

type ComplexIntResponse string

func (s ComplexIntResponse) ToUInt64() uint64 {

	sResult, _ := strconv.ParseUint(string(s), 16, 64)
	return sResult

}

func (s ComplexIntResponse) ToInt64() int64 {

	big, _ := new(big.Int).SetString(strings.TrimPrefix(string(s), "0x"), 16)
	return big.Int64()

}

func (s ComplexIntResponse) ToBigInt() *big.Int {
	big, _ := new(big.Int).SetString(string(s), 16)
	return big
}
