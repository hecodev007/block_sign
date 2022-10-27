package util

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// ParseInt parse hex string value to int
func ParseInt(value string) (int, error) {
	i, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func ParseUint(value string) (uint, error) {
	i, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		return 0, err
	}

	return uint(i), nil
}

func ParseUint64(value string) (uint64, error) {
	i, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		return 0, err
	}

	return uint64(i), nil
}

func ParseInt64(value string) (int64, error) {
	i, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return 0, err
	}

	return int64(i), nil
}

// ParseBigInt parse hex string value to big.Int
func ParseBigInt(value string) (*big.Int, error) {
	i := &big.Int{}
	_, err := fmt.Sscan(value, i)

	return i, err
}

// IntToHex convert int to hexadecimal representation
func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}
