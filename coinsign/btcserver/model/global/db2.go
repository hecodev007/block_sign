package global

var (
	KeyStore2 = make(map[string][]byte)
)

func GetValue2(address string) (string, bool) {
	_, ok := KeyStore2[address]
	if !ok {
		return "", false
	}
	d := KeyStore2[address]
	return string(d), true
}

func SetValue2(address, privKey string) {
	KeyStore2[address] = []byte(privKey)
}

func DelKeys2(address string) {
	delete(KeyStore2, address)
}
