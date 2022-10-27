package global

import (
	"sync"
)

var (
	KeyStore = sync.Map{}
)

func GetValue(keyName string) (string, bool) {
	v, ok := KeyStore.Load(keyName)
	if !ok {
		return "", false
	}
	d, _ := v.([]byte)
	return string(d), true
}

func SetValue(key, value string) {
	KeyStore.Store(key, []byte(value))
}

func DelKeys(key string) {
	KeyStore.Delete(key)
}
