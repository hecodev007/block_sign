
package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type ComplexString string

func (s ComplexString) ToHex() string {
	if strings.HasPrefix(string(s), "0x") {
		return string(s)
	}
	return fmt.Sprintf("0x%x", s)

}

func (s ComplexString) ToString() string {

	stringValue := string(s)

	sResult, _ := hex.DecodeString(strings.TrimPrefix(stringValue, "0x"))

	return s.clean(string(sResult))

}

func (s ComplexString) clean(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 32 && c < 127 {
			b[bl] = c
			bl++
		}
	}
	return strings.TrimSpace(string(b[:bl]))
}
