package global

import (
	"github.com/group-coldwallet/bchsign/secure"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	KeyStore = sync.Map{}
)


func GetValue(keyName string) (string, bool) {
	logrus.Infof("address %s need to get private key", keyName)
	key, err := secure.GetPrivateKey(keyName)
	if err != nil {
		logrus.Errorf("secure.GetPrivateKey error: %v", err)
		return "", false
	}
	return key, true
}

func SetValue(key, value string) {
	KeyStore.Store(key, []byte(value))
}

func DelKeys(key string) {
	KeyStore.Delete(key)
}
