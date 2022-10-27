package global

var (
	KeyStore = make(map[string][]byte)
)

func GetValue(address string) (string, bool) {
	_, ok := KeyStore[address]
	if !ok {
		return "", false
	}
	d := KeyStore[address]
	return string(d), true
}

func SetValue(address, privKey string) {
	KeyStore[address] = []byte(privKey)
}

func DelKeys(address string) {
	delete(KeyStore, address)
}
