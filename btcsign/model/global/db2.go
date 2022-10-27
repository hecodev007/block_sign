package global

import (
	"github.com/group-coldwallet/btcsign/secure"
	"github.com/sirupsen/logrus"
)

var (
	KeyStore2 = make(map[string][]byte)
)

func GetValue2(mch string, keyName string) (string, bool) {
	logrus.Infof("address %s need to get private key", keyName)
	key, err := secure.GetPrivateKey(mch, keyName)
	if err != nil {
		logrus.Errorf("secure.GetPrivateKey error: %v", err)
		return "", false
	}
	return key, true
}

func SetValue2(address, privKey string) {
	KeyStore2[address] = []byte(privKey)
}

func DelKeys2(address string) {
	delete(KeyStore2, address)
}
