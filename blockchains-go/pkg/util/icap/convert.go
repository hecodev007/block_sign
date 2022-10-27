package icap

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strconv"
	"strings"
)

var (
	Base36Chars        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	errICAPLength      = errors.New("invalid ICAP length")
	errICAPEncoding    = errors.New("invalid ICAP encoding")
	errICAPChecksum    = errors.New("invalid ICAP checksum")
	errICAPCountryCode = errors.New("invalid ICAP country code")
	errICAPAssetIdent  = errors.New("invalid ICAP asset identifier")
	errICAPInstCode    = errors.New("invalid ICAP institution code")
	errICAPClientIdent = errors.New("invalid ICAP client identifier")
)

var (
	Big1  = big.NewInt(1)
	Big0  = big.NewInt(0)
	Big36 = big.NewInt(36)
	Big97 = big.NewInt(97)
	Big98 = big.NewInt(98)
)

//export ConvertICAPToAddress
func ConvertICAPToAddress(s string) (common.Address, error) {
	switch len(s) {
	case 35:
		return parseICAP(s)
	case 34:
		return parseICAP(s)
	case 20:
		return parseIndirectICAP(s)
	default:
		return common.Address{}, errICAPLength
	}
}

func ConvertICAPToAddressRon(s string) (common.Address, error) {
	if strings.HasPrefix(s, "ronin:") {
		addr := "0x" + s[6:]
		return common.HexToAddress(addr), nil
	}
	return common.Address{}, errors.New("error addr")
}

func parseICAP(s string) (common.Address, error) {
	if !strings.HasPrefix(s, "TH") {
		return common.Address{}, errICAPCountryCode
	}
	if err := validCheckSum(s); err != nil {
		return common.Address{}, err
	}
	bigAddr, _ := new(big.Int).SetString(s[4:], 36)
	return common.BigToAddress(bigAddr), nil
}

func parseIndirectICAP(s string) (common.Address, error) {
	if !strings.HasPrefix(s, "TH") {
		return common.Address{}, errICAPCountryCode
	}
	if s[4:7] != "ETH" {
		return common.Address{}, errICAPAssetIdent
	}
	if err := validCheckSum(s); err != nil {
		return common.Address{}, err
	}
	return common.Address{}, errors.New("not implemented")
}

func ConvertAddressToICAP(a common.Address) (string, error) {
	enc := base36Encode(common.HexToHash(a.Hex()).Big())
	if len(enc) < 30 {
		enc = join(strings.Repeat("0", 30-len(enc)), enc)
	}
	icap := join("TH", checkDigits(enc), enc)
	return icap, nil
}

func validCheckSum(s string) error {
	s = join(s[4:], s[:4])
	expanded, err := iso13616Expand(s)
	if err != nil {
		return err
	}
	checkSumNum, _ := new(big.Int).SetString(expanded, 10)
	if checkSumNum.Mod(checkSumNum, Big97).Cmp(Big1) != 0 {
		return errICAPChecksum
	}
	return nil
}

func checkDigits(s string) string {
	expanded, _ := iso13616Expand(strings.Join([]string{s, "TH00"}, ""))
	num, _ := new(big.Int).SetString(expanded, 10)
	num.Sub(Big98, num.Mod(num, Big97))

	checkDigits := num.String()
	if len(checkDigits) == 1 {
		checkDigits = join("0", checkDigits)
	}
	return checkDigits
}

func iso13616Expand(s string) (string, error) {
	var parts []string
	if !validBase36(s) {
		return "", errICAPEncoding
	}
	for _, c := range s {
		i := uint64(c)
		if i >= 65 {
			parts = append(parts, strconv.FormatUint(uint64(c)-55, 10))
		} else {
			parts = append(parts, string(c))
		}
	}
	return join(parts...), nil
}

func base36Encode(i *big.Int) string {
	var chars []rune
	x := new(big.Int)
	for {
		x.Mod(i, Big36)
		chars = append(chars, rune(Base36Chars[x.Uint64()]))
		i.Div(i, Big36)
		if i.Cmp(Big0) == 0 {
			break
		}
	}
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func validBase36(s string) bool {
	for _, c := range s {
		i := uint64(c)
		if i < 48 || (i > 57 && i < 65) || i > 90 {
			return false
		}
	}
	return true
}

func join(s ...string) string {
	return strings.Join(s, "")
}
