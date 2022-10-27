package db

import (
	"sync"
)

var (
	KeyStore = sync.Map{}
)

func GetPrivKey(pubKey string) (string, bool) {
	v, ok := KeyStore.Load(pubKey)
	if !ok {
		return "", false
	}
	d, _ := v.([]byte)
	return string(d), true
}

func SetKeys(pubKey, privKey string) {
	KeyStore.Store(pubKey, []byte(privKey))
}

func DelKeys(pubKey string) {
	KeyStore.Delete(pubKey)
}
