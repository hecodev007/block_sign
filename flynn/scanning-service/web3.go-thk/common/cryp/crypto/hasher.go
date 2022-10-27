package crypto

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/common/encoding"
	"reflect"
)

var (
	NilHashSlice, _ = Hash256s(nil)
)

func HashObject(o interface{}) ([]byte, error) {
	if o == nil {
		return NilHashSlice, nil
	}
	v := reflect.ValueOf(o)
	if !v.IsValid() {
		return NilHashSlice, nil
	}

	switch val := o.(type) {
	case Hasher:
		return val.HashValue()
	default:
		hasher := GetHash256()
		if err := encoding.Encode(val, hasher); err != nil {
			return nil, err
		}
		return hasher.Sum(nil), nil
	}
}
