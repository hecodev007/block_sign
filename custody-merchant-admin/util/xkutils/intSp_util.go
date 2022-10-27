package xkutils

import (
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

func Int64SplitByString(s string, split string) ([]int64, error) {
	var uids []int64
	sp := strings.Split(s, split)
	for i := 0; i < len(sp); i++ {
		atoi, err := strconv.Atoi(sp[i])
		if err != nil {
			return nil, err
		}
		uids = append(uids, int64(atoi))
	}
	return uids, nil
}

func IntSplitByString(s string, split string) ([]int, error) {
	var uids []int
	sp := strings.Split(s, split)
	for i := 0; i < len(sp); i++ {
		atoi, err := strconv.Atoi(sp[i])
		if err != nil {
			return nil, err
		}
		uids = append(uids, atoi)
	}
	return uids, nil
}

func StrToInt(s string) int {
	if s != "" {
		atoi, err := strconv.Atoi(s)
		if err != nil {
			return 0
		} else {
			return atoi
		}
	}
	return 0
}

func StrToInt64(s string) int64 {
	if s != "" {
		atoi, err := strconv.ParseInt(s, 0, 40)
		if err != nil {
			return 0
		} else {
			return atoi
		}
	}
	return 0
}

func StrToFloat64(s string) float64 {
	if s != "" {
		f, err := strconv.ParseFloat(s, 40)
		if err != nil {
			return 0.0
		} else {
			return f
		}
	}
	return 0.0
}

func IntToString(i int) string {
	if i != 0 {
		return strconv.Itoa(i)
	}
	return "0"
}

func IntListToStrList(list []int) []string {
	slist := []string{}
	for _, i := range list {
		slist = append(slist, strconv.Itoa(i))
	}
	return slist
}

func StringToDecimal(str string) decimal.Decimal {
	if str == "" {
		return decimal.Zero
	} else {
		fromString, err := decimal.NewFromString(str)
		if err != nil {
			return decimal.Zero
		}
		return fromString
	}
	return decimal.Zero
}
