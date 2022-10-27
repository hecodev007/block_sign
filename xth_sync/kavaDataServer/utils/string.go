package utils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
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

func Int64ToHex(i int64) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}

// utc时间转换成zh字符串时间
func TimeToStr(val int64) string {
	if val == 0 {
		return "2006-01-02 15:04:05"
	}
	tm := time.Unix(val, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// utc时间转换成zh字符串时间
func StrToTime(val string) int64 {
	if val == "" {
		return 0
	}
	p, _ := time.Parse("2006-01-02 15:04:05", val)
	return p.Unix()
}
